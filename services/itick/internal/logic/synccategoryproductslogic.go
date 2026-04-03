package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncCategoryProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncCategoryProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncCategoryProductsLogic {
	return &SyncCategoryProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步类型下的产品
func (l *SyncCategoryProductsLogic) SyncCategoryProducts(in *itick.SyncCategoryProductsReq) (*itick.SyncCategoryProductsResp, error) {
	result, err := l.svcCtx.ItickCategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		logx.Errorf("find category failed, err=%v", err)
		return &itick.SyncCategoryProductsResp{
			Base: &itick.RespBase{
				Code: 1,
				Msg:  err.Error(),
			},
		}, nil
	}
	if result == nil {
		return &itick.SyncCategoryProductsResp{
			Base: &itick.RespBase{
				Code: 1,
				Msg:  "分类不存在",
			},
		}, nil
	}

	taskNo := fmt.Sprintf("sync_category_products_%d_%d", in.Id, time.Now().UnixNano())
	now := time.Now().UnixMilli()

	// 创建任务记录
	_, err = l.svcCtx.ItickSyncTaskModel.Insert(l.ctx, &models.TItickSyncTask{
		TaskNo:    taskNo,
		TaskType:  "sync_category_products",
		BizId:     in.Id,
		Status:    0,
		Message:   "任务已提交",
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		logx.Errorf("create sync task failed, err=%v", err)
		return &itick.SyncCategoryProductsResp{
			Base: &itick.RespBase{
				Code: 1,
				Msg:  "创建同步任务失败",
			},
		}, nil
	}

	// 拷贝参数，避免直接引用请求对象
	reqCopy := *in

	// 后台异步执行
	go func(taskNo string, reqCopy *itick.SyncCategoryProductsReq) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		logic := NewSyncCategoryProductsWorker(bgCtx, l.svcCtx)
		logic.Run(taskNo, reqCopy)
	}(taskNo, &reqCopy)

	return &itick.SyncCategoryProductsResp{
		Base: &itick.RespBase{
			Code: 0,
			Msg:  "同步任务已提交",
		},
		TaskNo: taskNo,
	}, nil
}

type SyncCategoryProductsWorker struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncCategoryProductsWorker(ctx context.Context, svcCtx *svc.ServiceContext) *SyncCategoryProductsWorker {
	return &SyncCategoryProductsWorker{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (w *SyncCategoryProductsWorker) Run(taskNo string, in *itick.SyncCategoryProductsReq) {
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("同步任务panic: %v", r)
			logx.Errorf(errMsg)
			_ = w.updateTaskStatus(taskNo, 3, errMsg)
		}
	}()

	_ = w.updateTaskStatus(taskNo, 1, "同步中")

	err := w.doSync(in)
	if err != nil {
		logx.Errorf("sync category products failed, taskNo=%s, err=%v", taskNo, err)
		_ = w.updateTaskStatus(taskNo, 3, err.Error())
		return
	}

	_ = w.updateTaskStatus(taskNo, 2, "同步成功")
}

func (w *SyncCategoryProductsWorker) doSync(in *itick.SyncCategoryProductsReq) error {
	result, err := w.svcCtx.ItickCategoryModel.FindOne(w.ctx, in.Id)
	if err != nil {
		return fmt.Errorf("find category failed: %w", err)
	}
	if result == nil {
		return errors.New("分类不存在")
	}

	regions, err := w.getRegion(result.CategoryCode)
	if err != nil {
		return fmt.Errorf("get region failed: %w", err)
	}

	for _, region := range regions {
		resp, err := w.getSymbolList(
			w.ctx,
			w.svcCtx.Config.Itick.ApiUrl,
			w.svcCtx.Config.Itick.Token,
			result.CategoryCode,
			region,
		)
		if err != nil {
			return fmt.Errorf("get symbol list failed, region=%s, err=%w", region, err)
		}

		for _, item := range resp.Data {
			_, err := w.svcCtx.ItickProductModel.Insert(w.ctx, &models.TItickProduct{
				CategoryType: result.CategoryType,
				Market:       region,
				Symbol:       item.Code,
				Code:         item.Code,
				Name:         item.Name,
				DisplayName:  item.Name,
				Exchange:     item.Exchange,
				Sector:       item.Sector,
				Lug:          item.Lug,
				BaseCoin:     "",
				QuoteCoin:    "",
				Enabled:      1,
				AppVisible:   1,
				Sort:         0,
				Icon:         "",
				Remark:       fmt.Sprintf("同步自 iTick，分类：%s，地区：%s", result.CategoryCode, region),
				CreateTime:   time.Now().UnixMilli(),
				UpdateTime:   time.Now().UnixMilli(),
			})
			if err != nil {
				logx.Errorf("insert product failed, code=%s, err=%v", item.Code, err)
				// 这里你自己决定：
				// 1. 遇到单条失败继续
				// 2. 直接终止整个任务
			}
		}
	}

	return nil
}

type SymbolListResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data []SymbolItem `json:"data"`
}

type SymbolItem struct {
	Code     string `json:"c"`           // 产品代码
	Name     string `json:"n"`           // 产品名称
	Type     string `json:"t"`           // stock/forex/indices/crypto/future/fund
	Exchange string `json:"e"`           // 交易所
	Sector   string `json:"s,omitempty"` // 行业/领域
	Lug      string `json:"l,omitempty"` // slug, URL友好标识
}

// GetSymbolList 调用 iTick Symbol List API。
// apiURL 例如: https://api.itick.org
// token  : iTick token
// category: stock/forex/indices/crypto/future/fund
// region : HK/US/BA/GB/CN ...
func (w *SyncCategoryProductsWorker) getSymbolList(ctx context.Context, apiURL, token, category, region string) (*SymbolListResponse, error) {
	apiURL = strings.TrimSpace(apiURL)
	token = strings.TrimSpace(token)
	category = strings.ToLower(strings.TrimSpace(category))
	region = strings.ToUpper(strings.TrimSpace(region))

	if apiURL == "" {
		return nil, fmt.Errorf("apiURL is required")
	}
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	if category == "" {
		return nil, fmt.Errorf("category is required")
	}
	if region == "" {
		return nil, fmt.Errorf("region is required")
	}

	switch category {
	case "stock", "forex", "indices", "crypto", "future", "fund":
	default:
		return nil, fmt.Errorf("unsupported category: %s", category)
	}

	base, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid apiURL: %w", err)
	}

	// 拼接 /symbol/list
	base.Path = path.Join(strings.TrimRight(base.Path, "/"), "/symbol/list")

	q := base.Query()
	q.Set("type", category)
	q.Set("region", region)

	// iTick 文档里 code 是必填。
	// 这里只做“按分类+地区”调用时，先给空字符串。
	// 如果你后面要精确搜索，建议把 code 作为方法参数补上。
	q.Set("code", "")
	base.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("token", token)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request itick symbol list failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("itick returned non-200 status: %d", resp.StatusCode)
	}

	var out SymbolListResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if out.Code != 0 {
		return &out, fmt.Errorf("itick business error: code=%d msg=%s", out.Code, out.Msg)
	}

	return &out, nil
}

func (lw *SyncCategoryProductsWorker) getRegion(category string) ([]string, error) {
	// 市场代码 股票包括（HK、SZ、SH、US、SG、JP、TW、IN、TH、DE、MX、MY、TR、ES、NL、GB、ID、VN），外汇（GB），指数（GB），数字币（BA）、期货（US、HK、CN）、基金（US）
	switch category {
	case "stock":
		return []string{"HK", "SZ", "SH", "US", "SG", "JP", "TW", "IN", "TH", "DE", "MX", "MY", "TR", "ES", "NL", "GB", "ID", "VN"}, nil
	case "forex":
		return []string{"GB"}, nil
	case "indices":
		return []string{"GB"}, nil
	case "crypto":
		return []string{"BA"}, nil
	case "future":
		return []string{"US", "HK", "CN"}, nil
	case "fund":
		return []string{"US"}, nil
	default:
		return nil, fmt.Errorf("unsupported category: %s", category)
	}
}

func (w *SyncCategoryProductsWorker) updateTaskStatus(taskNo string, status int64, message string) error {
	return w.svcCtx.ItickSyncTaskModel.UpdateStatusByTaskNo(w.ctx, taskNo, status, message, time.Now().UnixMilli())
}
