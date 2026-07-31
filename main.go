package main

import "net/http"
import "sync/atomic"
import "fmt"

type apiConfig struct {
    fileserverHits atomic.Int32
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
func (cfg *apiConfig) fileserverHitsHandler(w http.ResponseWriter, r *http.Request) {

       w.Header().Set("Content-Type", "text/html; charset=utf-8")
       w.WriteHeader(http.StatusOK)
       w.Write([]byte(fmt.Sprintf(`<html>
         <body>
           <h1>Welcome, Chirpy Admin</h1>
           <p>Chirpy has been visited %d times!</p>
         </body>
       </html>`, cfg.fileserverHits.Load())))
   }
func (cfg *apiConfig) resetFileserverHitsHandler(w http.ResponseWriter, r *http.Request) {
         cfg.fileserverHits.Store(0)
         w.Header().Set("Content-Type", "text/plain; charset=utf-8")
         w.WriteHeader(http.StatusOK)
         w.Write([]byte("Hits counter reset"))
    }

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.Handle("/api/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", apiCfg.fileserverHitsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetFileserverHitsHandler)


	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	srv.ListenAndServe()
}
