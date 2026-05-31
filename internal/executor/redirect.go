//go:build !js

package executor

import "net/http"

func redirectRecorder(req *http.Request, via []*http.Request) error {
	if rv := req.Context().Value(ctxKeyTimingRecorder{}); rv != nil {
		if tr, ok := rv.(*TimingRecorder); ok {
			hop := RedirectHop{HopNumber: len(via), SourceURL: "", TargetURL: req.URL.String(), StatusCode: 0}
			if len(via) > 0 {
				hop.SourceURL = via[len(via)-1].URL.String()
			}
			tr.Redirects = append(tr.Redirects, hop)
		}
	}
	return nil
}
