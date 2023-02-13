package core

import (
	"log"
	"net/http"

	"github.com/google/uuid"
)

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func Cros(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// 跨域名
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func GetTokenStringFromRequest(r *http.Request) string {
	token := r.Header.Get("Authorization")
	return token
}

// 鉴权
func Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 这里

		token := GetTokenStringFromRequest(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}
		if _, err := ParseToken(token); err != nil {
			log.Println(r.URL.Path, err.Error())
			return
			//next.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}

	})
}

// cookie 操作
// 这个下面其实是不需要的
func Cookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const sessionKey = "SESSIONID"
		if ssId, err := r.Cookie(sessionKey); err != nil {
			next.ServeHTTP(w, r)
		} else {
			if ssId.Value == "" {
				ck := &http.Cookie{
					Name:  sessionKey,
					Value: uuid.New().String(),
					Path:  "/",
				}
				http.SetCookie(w, ck)
			}
			next.ServeHTTP(w, r)
		}
	})
}
