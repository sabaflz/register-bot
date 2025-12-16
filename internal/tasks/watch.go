package tasks

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func (t *Task) CheckEnrollmentData(CRN string) error {
	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
	}

	values := url.Values{
		"term":                  {t.TermID},
		"courseReferenceNumber": {CRN},
	}

	response, err := t.DoReq(t.MakeReq("POST", "https://reg.oci.fhda.edu/StudentRegistrationSsb/ssb/searchResults/getEnrollmentInfo", headers, []byte(values.Encode())), fmt.Sprintf("Getting Enrollment Data (%s)", CRN), true)
	if err != nil {
		fmt.Println(err)
		if response != nil {
			discardResp(response)
		}
		return err
	}

	if response == nil {
		return fmt.Errorf("received nil response")
	}

	body, _ := readBody(response)
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		discardResp(response)
		return err
	}

	var enrollmentSeatsAvailable, waitlistCapacity, waitlistActual, waitlistSeatsAvailable string

	document.Find("span.status-bold").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Enrollment Seats Available:") {
			enrollmentSeatsAvailable = s.Next().Text()
		} else if strings.Contains(s.Text(), "Waitlist Seats Available:") {
			waitlistSeatsAvailable = s.Next().Text()
		} else if strings.Contains(s.Text(), "Waitlist Capacity:") {
			waitlistCapacity = s.Next().Text()
		} else if strings.Contains(s.Text(), "Waitlist Actual:") {
			waitlistActual = s.Next().Text()
		}
	})

	numEnrollmentSeatsAvailable, _ := strconv.Atoi(enrollmentSeatsAvailable)
	numWaitlistCapacity, _ := strconv.Atoi(waitlistCapacity)
	numWaitlistActual, _ := strconv.Atoi(waitlistActual)
	numWaitlistSeatsAvailable, _ := strconv.Atoi(waitlistSeatsAvailable)

	// Auto-enroll if ANY seat is available (enrollment OR waitlist)
	hasEnrollmentSeat := numEnrollmentSeatsAvailable > 0
	hasWaitlistSeat := numWaitlistCapacity > numWaitlistActual && numWaitlistSeatsAvailable > 0

	if hasEnrollmentSeat || hasWaitlistSeat {
		var message string
		if hasEnrollmentSeat {
			message = fmt.Sprintf("[%s] %s Enrollment seat(s) is now Available - Auto-enrolling!", CRN, enrollmentSeatsAvailable)
			t.WaitlistTask = false
		} else {
			message = fmt.Sprintf("[%s] %s Waitlist spot(s) is now Available - Auto-enrolling!", CRN, waitlistSeatsAvailable)
			t.WaitlistTask = true
		}
		
		t.SendNotification("Watch Task - Seat Available", message)
		fmt.Println(message)
		t.CRNs = []string{CRN}
		t.Signup()
		return nil
	} else {
		if numEnrollmentSeatsAvailable >= 1 && numWaitlistSeatsAvailable == 0 {
			fmt.Printf("[%s] - (Waitlist Opening Soon)\n", CRN)
		} else {
			fmt.Printf("[%s] - (Not Available)\n", CRN)
		}
		time.Sleep(5 * time.Second)
		return t.CheckEnrollmentData(CRN)
	}
}

func (t *Task) Watch() error {

	var waitGroup sync.WaitGroup
	errChan := make(chan error, len(t.CRNs))

	for _, course := range t.CRNs {
		waitGroup.Add(1)

		go func(course string) {
			defer waitGroup.Done()
			if err := t.CheckEnrollmentData(course); err != nil {
				errChan <- err
			}
		}(course)
	}

	waitGroup.Wait()
	close(errChan)
	return nil
}
