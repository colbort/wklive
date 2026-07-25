package callback

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"wklive/payment-api/internal/svc"
	"wklive/payment-api/internal/types"
)

func handleNotify(
	w http.ResponseWriter,
	r *http.Request,
	svcCtx *svc.ServiceContext,
	req *types.NotifyReq,
) *types.NotifyReq {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, svcCtx.Config.MaxBytes))
	if err != nil {
		http.Error(w, "invalid notification body", http.StatusBadRequest)
		return nil
	}
	headers := make(map[string]string, len(r.Header))
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	query := make(map[string]string, len(r.URL.Query()))
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			query[key] = values[0]
		}
	}
	form := make(map[string]string)
	if strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/x-www-form-urlencoded") {
		if values, parseErr := url.ParseQuery(string(body)); parseErr == nil {
			form = make(map[string]string, len(values))
			for key, items := range values {
				if len(items) > 0 {
					form[key] = items[0]
				}
			}
		}
	}
	req.Headers = headers
	req.Query = query
	req.Form = form
	req.Body = body
	return req
}
