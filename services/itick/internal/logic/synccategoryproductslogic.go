package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/pkg/utils"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/gogo/protobuf/proto"
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

// 同步类型下的产品 (管理后台手动)
func (l *SyncCategoryProductsLogic) SyncCategoryProducts(in *itick.SyncCategoryProductsReq) (*itick.SyncCategoryProductsResp, error) {
	result, err := l.svcCtx.ItickCategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		logx.Errorf("find category failed, err=%v", err)
		return &itick.SyncCategoryProductsResp{
			Base: helper.GetErrResp(i18n.CategoryNotFound, i18n.Translate(i18n.CategoryNotFound, l.ctx)),
		}, nil
	}
	if result == nil {
		return &itick.SyncCategoryProductsResp{
			Base: helper.GetErrResp(i18n.CategoryNotFound, i18n.Translate(i18n.CategoryNotFound, l.ctx)),
		}, nil
	}

	taskNo := fmt.Sprintf("sync_category_products_%d_%d", in.Id, time.Now().UnixNano())
	now := cutils.NowMillis()

	// 创建任务记录
	_, err = l.svcCtx.ItickSyncTaskModel.Insert(l.ctx, &models.TItickSyncTask{
		TaskNo:      taskNo,
		TaskType:    "sync_category_products",
		BizId:       in.Id,
		Status:      0,
		Message:     "任务已提交",
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		logx.Errorf("create sync task failed, err=%v", err)
		return &itick.SyncCategoryProductsResp{
			Base: helper.GetErrResp(i18n.SyncTaskCreateFailed, i18n.Translate(i18n.SyncTaskCreateFailed, l.ctx)),
		}, nil
	}

	// 拷贝参数，避免直接引用请求对象
	reqCopy := proto.Clone(in).(*itick.SyncCategoryProductsReq)

	// 后台异步执行
	go func(taskNo string, reqCopy *itick.SyncCategoryProductsReq) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		logic := NewSyncCategoryProductsWorker(bgCtx, l.svcCtx)
		logic.Run(taskNo, reqCopy)
	}(taskNo, reqCopy)

	return &itick.SyncCategoryProductsResp{
		Base:   helper.OkResp(),
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
			logx.Errorf("%s", errMsg)
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
		return i18n.StatusError(w.ctx, i18n.InternalServerError)
	}
	if result == nil {
		return i18n.StatusError(w.ctx, i18n.CategoryNotFound)
	}

	regions, err := w.getRegion(result.CategoryCode)
	if err != nil {
		return i18n.StatusError(w.ctx, i18n.MarketRequired)
	}

	for _, market := range regions {
		resp, err := w.getSymbolList(
			w.ctx,
			w.svcCtx.Config.Itick.ApiUrl,
			w.svcCtx.Config.Itick.Token,
			result.CategoryCode,
			market,
		)
		if err != nil {
			return err
		}

		for _, item := range resp.Data {
			_, err := w.svcCtx.ItickProductModel.Upsert(w.ctx, &models.TItickProduct{
				CategoryType: result.CategoryType,
				CategoryName: result.CategoryName,
				CategoryCode: result.CategoryCode,
				Market:       market,
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
				Remark:       fmt.Sprintf("同步自 iTick，分类：%s，地区：%s", result.CategoryCode, market),
				CreateTimes:  cutils.NowMillis(),
				UpdateTimes:  cutils.NowMillis(),
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
// market : HK/US/BA/GB/CN ...
func (w *SyncCategoryProductsWorker) getSymbolList(ctx context.Context, apiURL, token, category, market string) (*SymbolListResponse, error) {
	apiURL = strings.TrimSpace(apiURL)
	token = strings.TrimSpace(token)
	category = strings.ToLower(strings.TrimSpace(category))
	market = strings.ToUpper(strings.TrimSpace(market))

	if apiURL == "" {
		return nil, i18n.StatusError(ctx, i18n.APIURLIsRequired)
	}
	if token == "" {
		return nil, i18n.StatusError(ctx, i18n.TokenRequired)
	}
	if category == "" {
		return nil, i18n.StatusError(ctx, i18n.CategoryRequired)
	}
	if market == "" {
		return nil, i18n.StatusError(ctx, i18n.MarketRequired)
	}

	if !utils.IsSupportedKlineCategory(category) {
		return nil, i18n.StatusError(ctx, i18n.CategoryNotFound)
	}

	base, err := url.Parse(apiURL)
	if err != nil {
		return nil, i18n.StatusError(ctx, i18n.ParamError)
	}

	// 拼接 /symbol/list
	base.Path = path.Join(strings.TrimRight(base.Path, "/"), "/symbol/list")

	q := base.Query()
	q.Set("type", category)
	q.Set("market", market)

	// iTick 文档里 code 是必填。
	// 这里只做“按分类+地区”调用时，先给空字符串。
	// 如果你后面要精确搜索，建议把 code 作为方法参数补上。
	q.Set("code", "")
	base.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, i18n.StatusError(ctx, i18n.InternalServerError)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("token", token)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, i18n.StatusError(ctx, i18n.ServiceUnavailable)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, i18n.StatusError(ctx, i18n.ServiceUnavailable)
	}

	var out SymbolListResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, i18n.StatusError(ctx, i18n.InternalServerError)
	}

	if out.Code != 0 {
		return &out, i18n.StatusError(ctx, i18n.ServiceUnavailable)
	}

	return &out, nil
}

func (lw *SyncCategoryProductsWorker) getRegion(category string) ([]string, error) {
	return utils.GetKlineCategoryRegions(category)
}

func (w *SyncCategoryProductsWorker) updateTaskStatus(taskNo string, status int64, message string) error {
	return w.svcCtx.ItickSyncTaskModel.UpdateStatusByTaskNo(w.ctx, taskNo, status, message, cutils.NowMillis())
}
