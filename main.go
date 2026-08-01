package main

import "net/http"
import "sync/atomic"
import "fmt"
import "encoding/json"


type apiConfig struct {
    fileserverHits atomic.Int32
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, err := json.Marshal(payload)
    if err != nil {
        w.WriteHeader(500)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
    type errorResponse struct {
        Error string `json:"error"`
    }
    respondWithJSON(w, code, errorResponse{Error: msg})
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

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
    type Chirp struct {
        Body string `json:"body"`
    }
    decoder := json.NewDecoder(r.Body)
    params := Chirp{}
    err := decoder.Decode(&params)
    if err != nil {
        respondWithError(w, 400, "Something went wrong")
        return
    }
    if len(params.Body) > 140 {
        respondWithError(w, 400, "Chirp is too long")
        return
    }
    type returnValues struct {
        Valid bool `json:"valid"`
    }
    respondWithJSON(w, 200, returnValues{Valid: true})
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
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)


	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	srv.ListenAndServe()
}
