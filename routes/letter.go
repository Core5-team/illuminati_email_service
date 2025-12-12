package routes

import (
	"encoding/json"
	"illuminati/go/microservice/utils"
	"log"
	"net/http"
	"os"
	"golang.org/x/sync/errgroup"
)

type participants struct {
	Participants []string `json:"participants"`
}

type LetterService struct {
	emailSender utils.EmailSender
	participantsURL string
}


func NewLetterService(emailSender utils.EmailSender, participantsURL string) *LetterService {
	return &LetterService{
		emailSender: emailSender,
		participantsURL : participantsURL,
	}
}

var(
	participants_url = os.Getenv("PARTICIPANTS_URL")
	ls = NewLetterService(utils.GetInstance(), participants_url)
)

type Letter struct {
	Topic        string   `json:"topic"`
	Text         string   `json:"text"`
	TargetEmails []string `json:"target_emails"`
}
func (ls *LetterService)getAppParticipants() ([]string, error) {

	resp, err := http.Get(ls.participantsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var data participants
	json.NewDecoder(resp.Body).Decode(&data)

	return data.Participants, nil
}

func (ls *LetterService) SendLetterEmail(request *http.Request) error {
	var g errgroup.Group
	var letter Letter
	var participants []string
	log.Print("participants url : ", ls.participantsURL)
	participants, err := ls.getAppParticipants()
	err = json.NewDecoder(request.Body).Decode(&letter)
	if err != nil {
		log.Println("Something went wrong while parsing the request body...")
		return err
	}

    g.Go(
		func() error {
		err = ls.emailSender.SendEmail(letter.Topic, letter.Text, participants)
		return err
		},
	)
	if err := g.Wait(); err != nil {
		log.Println("Something went wrong while sending the emails...")
		return err
	}

	return nil
}

func (ls *LetterService)PostLetter(w http.ResponseWriter, r *http.Request) {

	err := ls.SendLetterEmail(r)
	if err != nil {
		log.Print("Error :", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}


