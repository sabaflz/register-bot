package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SignupSession struct {
	SAMLRequest string
	Models      []map[string]interface{}
}

func (t *Task) CheckAuthSession() error {

	headers := [][2]string{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		{"accept-language", "en-US,en;q=0.9"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}

	response, err := t.DoReq(t.MakeReq("GET", "https://reg.oci.fhda.edu/StudentRegistrationSsb/login/authAjax", headers, nil), "Checking Auth Session", true)
	if err != nil {
		discardResp(response)
		return err
	}
	body, _ := readBody(response)
	if strings.Contains(string(body), "userNotLoggedIn") {
		t.GenSession()
	}
	return nil
}

func (t *Task) RegisterPostSignIn() error {
	headers := [][2]string{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		{"accept-language", "en-US,en;q=0.9"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}

	response, err := t.DoReq(t.MakeReq("GET", "https://reg.oci.fhda.edu/StudentRegistrationSsb/ssb/registration/registerPostSignIn?mode=registration", headers, nil), "Register Post Sign In", true)
	if err != nil {
		discardResp(response)
		return err
	}
	body, _ := readBody(response)
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		discardResp(response)
		return err
	}
	t.Session.SignupSession.SAMLRequest = getSelectorAttr(document, "input[name='SAMLRequest']", "value")
	return nil
}

func (t *Task) SubmitSamIsso() error {

	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}

	values := url.Values{
		"SAMLRequest": {t.Session.SignupSession.SAMLRequest},
	}

	response, err := t.DoReq(t.MakeReq("POST", "https://eis-prod.ec.fhda.edu/samlsso", headers, []byte(values.Encode())), "Submitting Sam Isso", true)
	if err != nil {
		discardResp(response)
		return err
	}

	body, _ := readBody(response)
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		discardResp(response)
		return err
	}
	t.Session.SAMLResponse = getSelectorAttr(document, "input[name='SAMLResponse']", "value")
	return nil
}

func (t *Task) SubmitSSBSp() error {
	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}

	values := url.Values{
		"SAMLResponse": {t.Session.SAMLResponse},
	}

	resp, err := t.DoReq(t.MakeReq("POST", "https://reg.oci.fhda.edu/StudentRegistrationSsb/saml/SSO/alias/registrationssb-prod-sp", headers, []byte(values.Encode())), "Submitting SSB SP", true)
	if err != nil {
		discardResp(resp)
		return err
	}
	return nil
}

func (t *Task) CheckCRN(course string) error {
	headers := [][2]string{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		{"accept-language", "en-US,en;q=0.9"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}

	response, err := t.DoReq(t.MakeReq("GET", fmt.Sprintf("https://reg.oci.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/getSectionDetailsFromCRN?courseReferenceNumber=%s&term=%s", course, t.TermID), headers, nil), fmt.Sprintf("Checking Course (%s)", course), true)
	if err != nil {
		discardResp(response)
		return err
	}
	body, _ := readBody(response)
	courseData := Course{}
	if err := json.Unmarshal(body, &courseData); err != nil {
		return err
	}
	if !courseData.Olr {
		fmt.Println(courseData.ResponseDisplay)
	} else {
		fmt.Printf("[%s] - Unable To Get Data\n", course)
	}
	return nil
}

func (t *Task) CheckCRNs() error {
	for _, course := range t.CRNs {
		t.CheckCRN(course)
	}
	return nil
}

func (t *Task) GetRegistrationStatus() error {
	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}

	values := url.Values{
		"term":            {t.TermID},
		"studyPath":       {},
		"startDatepicker": {},
		"endDatepicker":   {},
		"uniqueSessionId": {t.Session.UniqueSessionId},
	}

	response, err := t.DoReq(t.MakeReq("POST", "https://reg.oci.fhda.edu/StudentRegistrationSsb/ssb/term/search?mode=registration", headers, []byte(values.Encode())), "Getting Registration Status", true)
	if err != nil {
		discardResp(response)
		return err
	}

	body, _ := readBody(response)
	registrationStatus := RegistrationStatus{}

	if err := json.Unmarshal(body, &registrationStatus); err != nil {
		return err
	}

	var hasFailure, hasRegistrationTime bool
	var timeFailure string

	for _, failure := range registrationStatus.StudentEligFailures {
		fmt.Println(failure)
		hasFailure = true
		if strings.Contains(failure, "You can register from") {
			hasRegistrationTime = true
			timeFailure = failure
			break
		}
	}

	if hasFailure && !hasRegistrationTime {
		return errors.New(registrationStatus.StudentEligFailures[len(registrationStatus.StudentEligFailures)])
	}

	if hasFailure && hasRegistrationTime {
		pattern := regexp.MustCompile(`\d{2}/\d{2}/\d{4} \d{2}:\d{2} [APM]{2}`)
		matches := pattern.FindAllString(timeFailure, -1)

		if len(matches) > 0 {
			location, _ := time.LoadLocation("America/Los_Angeles")
			targetTime, _ := time.ParseInLocation("01/02/2006 03:04 PM", matches[0], location)
			now := time.Now().In(location)
			saveRegistrationTime(matches[0])

			if now.After(targetTime) {
				time.Sleep(2 * time.Second)
				return t.GetRegistrationStatus()
			} else if now.Before(targetTime) {
				t.CheckCRNs()
				timeToWait := targetTime.Sub(now)

				fmt.Printf("Waiting for Registration to open: %s\n", targetTime.Format(time.RFC1123))
				fmt.Printf("Will continue in %s\n", formatDuration(targetTime.Sub(now)))
				go func() {
					ticker := time.NewTicker(5 * time.Minute)
					defer ticker.Stop()

					endTime := time.Now().Add(timeToWait)
					for now := range ticker.C {
						err := t.CheckAuthSession()
						if err != nil {
							fmt.Println(err)
						}

						if now.After(endTime) {
							break
						}
					}
				}()

				time.Sleep(targetTime.Sub(now))
				err := t.CheckAuthSession()
				if err != nil {
					return err
				}
				return t.GetRegistrationStatus()
			}
		}
	}
	return nil
}

func (t *Task) VisitClassRegistration() error {
	headers := [][2]string{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		{"accept-language", "en-US,en;q=0.9"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}
	response, err := t.DoReq(t.MakeReq("HEAD", "https://reg.oci.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/classRegistration", headers, nil), "Visiting Class Registration", true)
	if err != nil {
		discardResp(response)
		return err
	}
	return nil
}

func (t *Task) AddCourse(course string) error {
	headers := [][2]string{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		{"accept-language", "en-US,en;q=0.9"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}

	response, err := t.DoReq(t.MakeReq("GET", fmt.Sprintf("https://reg.oci.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/addRegistrationItem?term=%s&courseReferenceNumber=%s&olr=false", t.TermID, course), headers, nil), fmt.Sprintf("Adding Course (%s)", course), true)
	if err != nil {
		discardResp(response)
		return err
	}
	body, _ := readBody(response)
	addCourse := AddCourse{}

	if err := json.Unmarshal(body, &addCourse); err != nil {
		return err
	}
	if addCourse.Success {
		model, err := extractModel([]byte(body))
		if err != nil {
			return err
		}
		// Use "WL" for waitlist if requested, otherwise "RW" for regular registration
		if t.WaitlistTask {
			model["selectedAction"] = "WL"
		} else {
			model["selectedAction"] = "RW"
		}
		t.Session.SignupSession.Models = append(t.Session.SignupSession.Models, model)
	} else {
		fmt.Printf("Error Adding Course (%s) - %s\n", course, addCourse.Message)
	}
	return nil
}

func (t *Task) DropCourse(course string) error {
	headers := [][2]string{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		{"accept-language", "en-US,en;q=0.9"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}

	// FHDA uses the same endpoint to "add" a course to the worksheet before submitting.
	// We need to get the model for the existing course to drop it.
	response, err := t.DoReq(t.MakeReq("GET", fmt.Sprintf("https://reg.oci.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/addRegistrationItem?term=%s&courseReferenceNumber=%s&olr=false", t.TermID, course), headers, nil), fmt.Sprintf("Preparing to Drop Course (%s)", course), true)
	if err != nil {
		discardResp(response)
		return err
	}
	body, _ := readBody(response)
	addCourse := AddCourse{}

	if err := json.Unmarshal(body, &addCourse); err != nil {
		return err
	}

	if addCourse.Success {
		model, err := extractModel([]byte(body))
		if err != nil {
			return err
		}
		model["selectedAction"] = "DW" // DW is typically the code for Web Drop
		t.Session.SignupSession.Models = append(t.Session.SignupSession.Models, model)
		fmt.Printf("Prepared to drop course %s\n", course)
	} else {
		fmt.Printf("Error preparing to drop course (%s) - %s\n", course, addCourse.Message)
	}
	return nil
}

func (t *Task) AddCourses() error {
	// First handle any drops
	for _, course := range t.DropCRNs {
		err := t.DropCourse(course)
		if err != nil {
			fmt.Printf("Warning: Failed to prepare drop for %s: %v\n", course, err)
		}
	}

	// Then handle adds
	for _, course := range t.CRNs {
		err := t.AddCourse(course)
		if err != nil {
			return err
		}
	}
	if len(t.Session.SignupSession.Models) == 0 {
		return errors.New("No courses to add or drop")
	}
	return nil
}

func (t *Task) SendBatch() error {
	headers := [][2]string{
		{"accept", "application/json"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/json"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}

	batch := Batch{
		Update:          t.Session.SignupSession.Models,
		UniqueSessionId: t.Session.UniqueSessionId,
	}

	batchJson, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		return nil
	}

	response, err := t.DoReq(t.MakeReq("POST", "https://reg.oci.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/submitRegistration/batch", headers, []byte(string(batchJson))), "Submitting Batch Update", true)
	if err != nil {
		discardResp(response)
		return err
	}
	body, _ := readBody(response)
	changes := Changes{}
	if err := json.Unmarshal(body, &changes); err != nil {
		return err
	}

	for _, data := range changes.Data.Update {
		// Check against both added and dropped CRNs
		allCRNs := append(t.CRNs, t.DropCRNs...)
		for _, courseReferenceNumber := range allCRNs {
			if data.CourseReferenceNumber == courseReferenceNumber {
				if data.StatusDescription == "Registered" {
					fmt.Printf("[%s - %s %s - %s] - Successfully Registered\n", data.CourseReferenceNumber, data.Subject, data.CourseNumber, data.CourseTitle)
					t.SendNotification(data.CourseTitle, fmt.Sprintf("Successful Enrollment (%s)", data.CourseReferenceNumber))
				} else if data.StatusDescription == "Waitlisted" {
					fmt.Printf("[%s - %s %s - %s] - Successfully Waitlisted\n", data.CourseReferenceNumber, data.Subject, data.CourseNumber, data.CourseTitle)
					t.SendNotification(data.CourseTitle, fmt.Sprintf("Successful Waitlisted (%s)", data.CourseReferenceNumber))
				} else if data.StatusDescription == "Deleted" || data.StatusDescription == "Dropped" || data.StatusDescription == "Web Drop" {
					fmt.Printf("[%s - %s %s - %s] - Successfully Dropped\n", data.CourseReferenceNumber, data.Subject, data.CourseNumber, data.CourseTitle)
					t.SendNotification(data.CourseTitle, fmt.Sprintf("Successful Drop (%s)", data.CourseReferenceNumber))
				} else if data.StatusDescription == "Errors Preventing Registration" {
					fmt.Printf("[%d] - Errors encountered processing [%s - %s %s - %s]\n", len(data.CrnErrors), data.CourseReferenceNumber, data.Subject, data.CourseNumber, data.CourseTitle)
					for _, err := range data.CrnErrors {
						fmt.Printf("%s\n", err.Message)
					}
				} else {
					fmt.Printf("[%s] Status: %s\n", data.CourseReferenceNumber, data.StatusDescription)
				}
			}
		}
	}
	return nil
}

func (t *Task) Signup() error {
	t.HomepageURL = "https://reg.oci.fhda.edu/StudentRegistrationSsb/saml/login"
	t.SSOManagerURL = "https://ssb-prod.ec.fhda.edu/ssomanager/saml/SSO"
	t.CheckAuthSession()
	if err := t.GetRegistrationStatus(); err != nil {
		return err
	}
	t.VisitClassRegistration()
	if err := t.AddCourses(); err != nil {
		return err
	}
	t.SendBatch()
	t.Client.CloseIdleConnections()
	return nil
}
