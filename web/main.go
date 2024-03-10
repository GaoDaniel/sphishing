package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

type StatsInner struct {
	Filename      string          `json:"filename"`
	Total         int             `json:"total"`
	Clicks        int             `json:"clicks"`
	RealismScores map[string]int  `json:"realismScores"`
	Codes         map[string]bool `json:"codes"`
}

type Stats struct {
	Stats StatsInner
	Lock  sync.Mutex
}

func (s *Stats) Clicks() (int, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if err := s.load(); err != nil {
		return 0, err
	}

	return s.Stats.Clicks, nil
}

func (s *Stats) Clickthrough() (float32, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	if err := s.load(); err != nil {
		return 0, err
	}
	return float32(s.Stats.Clicks) / float32(s.Stats.Total), nil
}

func generateCodes(n int) []string {
	codes := make([]string, n)
	for i := 0; i < n; i++ {
		codes[i] = uuid.New().String()
	}
	return codes
}

// create new Stats struct stored in filename
func NewStats(filename string, total int) (*Stats, error) {
	s := &Stats{
		Stats: StatsInner{
			Filename:      filename,
			Clicks:        0,
			Total:         total,
			RealismScores: make(map[string]int),
			Codes:         make(map[string]bool),
		},
	}

	codes := generateCodes(total)

	for _, code := range codes {
		s.Stats.RealismScores[code] = -1
		s.Stats.Codes[code] = true
	}

	if err := s.flush(); err != nil {
		return nil, err
	}

	return s, nil
}

// write stats to disk
// must hold lock!
func (s *Stats) flush() error {
	serialized, err := json.Marshal(&s.Stats)
	if err != nil {
		return err
	}

	if err = os.WriteFile(s.Stats.Filename, serialized, 0644); err != nil {
		return err
	}

	return nil
}

// must hold lock!
func (s *Stats) load() error {
	data, err := os.ReadFile(s.Stats.Filename)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, &s.Stats); err != nil {
		return err
	}

	return nil
}

// returns whether successfully redeemed
func (s *Stats) Redeem(code string) (bool, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if err := s.load(); err != nil {
		return false, err
	}

	valid, present := s.Stats.Codes[code]

	if !present {
		return false, nil
	} else if valid {
		// redeem the code
		s.Stats.Codes[code] = false
		s.Stats.Clicks++

		if err := s.flush(); err != nil {
			return false, err
		}

		return true, nil
	}

	// either code does not exist or it has been redeemed
	return false, nil
}

// realism score
func (s *Stats) Score(code string) (int, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if err := s.load(); err != nil {
		return 0, err
	}

	score, present := s.Stats.RealismScores[code]
	if !present {
		return 0, errors.New("code does not exist")
	}

	return score, nil
}

func (s *Stats) SetScore(code string, score int) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if err := s.load(); err != nil {
		return err
	}

	_, present := s.Stats.RealismScores[code]
	if !present {
		return errors.New("code does not exist")
	}

	s.Stats.RealismScores[code] = score

	if err := s.flush(); err != nil {
		return err
	}

	return nil

}

func main() {
	// TODO: in actual deployment, make two stats objects (one for finetune, one for base)
	stats, err := NewStats("codes.json", 40)
	if err != nil {
		panic("failed to create stats struct")
	}

	tmpl := template.Must(template.ParseFiles("rate.html"))

	http.HandleFunc("/rate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		code := r.URL.Query().Get("code")

		if r.Method == http.MethodGet {
			tmpl.Execute(w, code)
		} else if r.Method == http.MethodPost {
			r.ParseForm()
			score, err := strconv.Atoi(r.Form.Get("score"))
			if err != nil {
				http.Error(w, "invalid score", http.StatusBadRequest)
			}
			if score >= 1 && score <= 5 {
				log.Printf("setting score for %s to %d...\n", code, score)
				err := stats.SetScore(code, score)
				if err != nil {
					http.Error(w, "error setting score", http.StatusBadRequest)
					log.Println(err.Error())
					return
				} else {
					fmt.Fprintln(w, "success!")
					return
				}
			} else {
				http.Error(w, "score must be between 1 and 5 inclusive", http.StatusBadRequest)
				return
			}
		}
	})

	http.HandleFunc("/redeem", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		code := r.URL.Query().Get("code")
		log.Printf("redeeming %s...\n", code)
		found, err := stats.Redeem(code)

		if err != nil {
			fmt.Fprintln(w, "failed to redeem")
		} else if !found {
			fmt.Fprintln(w, "code does not exist or has been redeemed already")
		} else {
			http.Redirect(w, r, "/rate?code="+code, http.StatusSeeOther)
		}
	})

	log.Println("starting server...")
	http.ListenAndServe(":8100", nil)
}
