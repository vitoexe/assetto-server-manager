package servermanager

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// viewChampionshipHandler shows details of a given Championship
func penaltyHandler(w http.ResponseWriter, r *http.Request) {
	remove, err := applyPenalty(r)

	if err != nil {
		AddErrFlashQuick(w, r, "Could not add/remove penalty")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	if remove {
		AddFlashQuick(w, r, "Penalty Removed!")
	} else {
		AddFlashQuick(w, r, "Penalty Added!")
	}

	http.Redirect(w, r, r.Referer(), http.StatusFound)
}

func applyPenalty(r *http.Request) (bool, error) {
	var results *SessionResults
	var remove bool
	var penaltyTime float64

	jsonFileName := chi.URLParam(r, "sessionFile")
	GUID := chi.URLParam(r, "driverGUID")

	results, err := LoadResult(jsonFileName + ".json")

	if err != nil {
		logrus.Errorf("could not load session result file, err: %s", err)
		return false, err
	}

	err = r.ParseForm()

	if err != nil {
		logrus.Errorf("could not load parse form, err: %s", err)
		return false, err
	}

	if r.FormValue("action") == "add" {
		penaltyString := r.FormValue("time-penalty")

		if penaltyString == "" {
			penaltyTime = 0
		} else {
			pen, err := strconv.ParseFloat(penaltyString, 64)

			if err != nil {
				logrus.Errorf("could not parse penalty time, err: %s", err)
				return false, err
			}

			penaltyTime = pen
		}
	} else {
		// remove penalty
		remove = true
	}

	for _, result := range results.Result {
		if result.DriverGUID == GUID {
			if remove {
				result.HasPenalty = false
				result.Disqualified = false
				result.PenaltyTime = 0
			} else {
				if penaltyTime == 0 {
					result.Disqualified = true
					result.HasPenalty = false
				} else {
					result.HasPenalty = true
					result.Disqualified = false

					timeParsed, err := time.ParseDuration(fmt.Sprintf("%.1fs", penaltyTime))

					if err != nil {
						logrus.Errorf("could not parse penalty time, err: %s", err)
						return false, err
					}

					result.PenaltyTime = timeParsed
				}
			}
		}
	}

	// sort results.Result, if disqualified go to back, if time penalty sort by laps completed then lap time
	sort.Slice(results.Result, func(i, j int) bool {
		if !results.Result[i].Disqualified && !results.Result[j].Disqualified {

			// if both drivers aren't disqualified
			if results.GetLaps(results.Result[i].DriverGUID) == results.GetLaps(results.Result[j].DriverGUID) {
				// if their number of laps are equal, compare lap times

				return results.GetTime(results.Result[i].TotalTime, results.Result[i].DriverGUID, true) <
							results.GetTime(results.Result[j].TotalTime, results.Result[j].DriverGUID, true)
			}

			return results.GetLaps(results.Result[i].DriverGUID) >= results.GetLaps(results.Result[j].DriverGUID)

		} else if results.Result[i].Disqualified && results.Result[j].Disqualified {

			// if both drivers ARE disqualified, compare their lap times / num laps
			if results.GetLaps(results.Result[i].DriverGUID) == results.GetLaps(results.Result[j].DriverGUID) {
				// if their number of laps are equal, compare lap times
				return results.GetTime(results.Result[i].TotalTime, results.Result[i].DriverGUID, true) <
					results.GetTime(results.Result[j].TotalTime, results.Result[j].DriverGUID, true)
			}

			return results.GetLaps(results.Result[i].DriverGUID) >= results.GetLaps(results.Result[j].DriverGUID)

		} else {
			// driver i is closer to the front than j if they are not disqualified and j is
			return results.Result[j].Disqualified
		}
	})

	path := filepath.Join(ServerInstallPath, "results", jsonFileName + ".json")

	file, err := os.Create(path)

	if err != nil {
		logrus.Errorf("could not load session result file, err: %s", err)
		return false, err
	}

	encoder := json.NewEncoder(file)

	err = encoder.Encode(results)

	if err != nil {
		logrus.Errorf("could not encode to session result file, err: %s", err)
		return false, err
	}

	return remove, nil
}
