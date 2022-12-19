package main

import (
	"encoding/json"
	"log"
	"os"
	"net/http"

	"github.com/xanzy/go-gitlab"
	authentication "k8s.io/api/authentication/v1beta1"
)

func main() {
	log.Println("GitLab Authn Webhook:", os.Getenv("GITLAB_API_ENDPOINT"))
	http.HandleFunc("/authenticate", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var tr authentication.TokenReview
		err := decoder.Decode(&tr)
		if err != nil {
			log.Println("[Error]", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"apiVersion": "authentication.k8s.io/v1beta1",
				"kind":       "TokenReview",
				"status": authentication.TokenReviewStatus{
					Authenticated: false,
				},
			})
			return
		}

		client, err := gitlab.NewClient(tr.Spec.Token, gitlab.WithBaseURL(os.Getenv("GITLAB_API_ENDPOINT")))
		if err != nil {
			log.Fatal(err)
		}

		// Get user
		user, _, err := client.Users.CurrentUser()
		if err != nil {
			log.Println("[Error]", err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"apiVersion": "authentication.k8s.io/v1beta1",
				"kind":       "TokenReview",
				"status": authentication.TokenReviewStatus{
					Authenticated: false,
				},
			})
			return
		}

		// Get user's group
    	lgo := gitlab.ListGroupsOptions{}
    	lgo.ListOptions.PerPage=100
		lgo.AllAvailable=gitlab.Bool(false)
		var all_group_path []string
		var count int = 0

		for {
			groups, resp, err := client.Groups.ListGroups(&lgo)
			if err != nil {
				log.Println("[Error]", err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"apiVersion": "authentication.k8s.io/v1beta1",
					"kind":       "TokenReview",
					"status": authentication.TokenReviewStatus{
						Authenticated: false,
					},
				})
				return
			}

			if len(all_group_path) == 0 {
				all_group_path = make([]string, resp.TotalItems)
			}

			for _, g := range groups {
				all_group_path[count] = g.FullPath
				count++
			}
	
			// Exit the loop when we've seen all pages.
			if resp.NextPage == 0 {
				break
			}
	
			// Update the page number to get the next page.
			lgo.Page = resp.NextPage
		}

		// Set the TokenReviewStatus
		log.Printf("[Success] login as %s, groups: %v", user.Username, all_group_path)
		w.WriteHeader(http.StatusOK)
		trs := authentication.TokenReviewStatus{
			Authenticated: true,
			User: authentication.UserInfo{
				Username: user.Username,
				UID:      user.Username,
				Groups: all_group_path,
			},
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"apiVersion": "authentication.k8s.io/v1beta1",
			"kind":       "TokenReview",
			"status":     trs,
		})
	})
	log.Fatal(http.ListenAndServe(":3000", nil))
}
