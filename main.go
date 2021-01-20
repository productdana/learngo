package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Article struct {
	Id      string `json:"Id"`
	Title   string `json:"Title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

var Articles = []Article{
	Article{Id: "1", Title: "Hello", Desc: "Article Description", Content: "Article Content"},
	Article{Id: "2", Title: "Hello 2", Desc: "Article Description", Content: "Article Content"},
}

func allArticles(w http.ResponseWriter, r *http.Request) {
	// articles := Articles{
	// 	Article{Id: "1", Title: "Test Title 1", Desc: "Test Description", Content: "Hello World"},
	// 	Article{Id: "2", Title: "Test Title 2", Desc: "Test Description", Content: "Hello World Again"},
	// }
	fmt.Println("endpoint hit: all articles endpoint")
	json.NewEncoder(w).Encode(Articles)
}

func returnSingleArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]

	for _, article := range Articles {
		if article.Id == key {
			json.NewEncoder(w).Encode(article)
		}
	}
}

func createNewArticle(w http.ResponseWriter, r *http.Request) {
	// get the body of the POST request
	// unmarshal this into a new Article struct
	// append this to our Articles array
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article Article
	json.Unmarshal(reqBody, &article)
	// update our global Articles array to include our new article
	Articles = append(Articles, article)
	json.NewEncoder(w).Encode(article)
	// fmt.Fprintf(w, "%+v", string(reqBody))
}

func deleteArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	for index, article := range Articles {
		if article.Id == key {
			Articles = append(Articles[:index], Articles[index+1:]...)
		}
	}
}

func updateArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	for index, article := range Articles {
		if article.Id == key {
			reqBody, _ := ioutil.ReadAll(r.Body)
			var updatedArticle Article
			json.Unmarshal(reqBody, &updatedArticle)
			// Is there an easy way to spread prev values and override only some properties?
			// combinedUpdatedArticle := {...Articles[index], ...updatedArticle}
			Articles[index] = updatedArticle
			json.NewEncoder(w).Encode(updatedArticle)
		}
		// how to return error if id not found?
	}
}

type Email struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// type Dadjoke struct {
// 	Id     string `json:"id"`
// 	Joke   string `json:"joke"`
// 	Status int    `json:"status"`
// }

type Dadjoke struct {
	Id     string
	Joke   string
	Status int
}

func getDadJoke() (string, error) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://icanhazdadjoke.com/", nil)

	req.Header.Add("Accept", "application/json") // Add or Set works

	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	resBody, _ := ioutil.ReadAll(res.Body)
	var dadjoke Dadjoke

	err = json.Unmarshal(resBody, &dadjoke)
	if err != nil {
		return "", err
	}

	return dadjoke.Joke, nil
}

func sendFunnyEmail(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var email Email
	json.Unmarshal(reqBody, &email)

	// fetch dad joke
	dadJoke, err := getDadJoke()
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(err)
		return
	}

	// attach dad joke at end of email body
	from := mail.NewEmail("funnypants", os.Getenv("TESTEMAIL"))
	subject := email.Subject + " - enhanced with dad joke"
	to := mail.NewEmail("funnypantsrecipient", os.Getenv("TESTEMAIL"))
	plainTextContent := email.Body + " \n \n dad joke: " + dadJoke
	htmlContent := "<strong>" + plainTextContent + "</strong>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))

	// send funny email
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
		w.WriteHeader(response.StatusCode)
		json.NewEncoder(w).Encode(err.Error)
	} else {
		// email sent
		fmt.Println(response.StatusCode)
		fmt.Println(response)
		json.NewEncoder(w).Encode(response)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage endpoint hit")
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/articles", allArticles)
	// order matters. this POST must be returned before the other /article endpoint
	myRouter.HandleFunc("/article", createNewArticle).Methods("POST")
	myRouter.HandleFunc("/article/{id}", deleteArticle).Methods("DELETE")
	myRouter.HandleFunc("/article/{id}", updateArticle).Methods("PATCH")
	myRouter.HandleFunc("/article/{id}", returnSingleArticle)

	myRouter.HandleFunc("/sendfunnyemail", sendFunnyEmail).Methods("POST")

	log.Fatal(http.ListenAndServe(":8081", myRouter))
}

// define entry point to app
func main() {
	handleRequests()
}
