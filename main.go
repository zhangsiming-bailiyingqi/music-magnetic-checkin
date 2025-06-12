package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// --- Configuration ---
// Move these out to environment variables or a config file in production.
const (
	signURL   = "https://www.hifini.com/sg_sign.htm"
	signValue = "7efe5c094ab965b636502ffeb502ca67821744dac0f53a4ef097e649ccd4fc1b"
	cookie    = "bbs_sid=l4o8n1tc2n0kjbjjlhoam0lt86; bbs_token=lHQRknttrC6cXUh6cpRSXMH2TCACfsRqtoJPJ3FoA_2FsffRAH3Ms_2B37opQ8SHtv5o98haFiXZ1_2B6pqI3JL82Nwp0WHbwauT_2BE"
)

// doSign performs the sign‑in request once and logs the result.
func doSign(ctx context.Context) error {
	form := url.Values{}
	form.Set("sign", signValue)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, signURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	// Mandatory headers copied from the original cURL command.
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "text/plain, */*; q=0.01")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://www.hifini.com")
	req.Header.Set("Referer", "https://www.hifini.com/")

	// *This cookie does not expire, so it can be used indefinitely.
	req.Header.Set("Cookie", cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[sign] status=%s body=%s", resp.Status, strings.TrimSpace(string(body)))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}
	return nil
}

// nextTick returns the duration until the next run at the given hour:minute (local time).
func nextTick(hour, minute int) time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !next.After(now) {
		// already passed today -> schedule for tomorrow
		next = next.Add(24 * time.Hour)
	}
	return time.Until(next)
}

func main() {
	// Choose one of the two scheduling strategies below.

	// * --- Strategy A: fire immediately, then every 24 h ---
	if err := doSign(context.Background()); err != nil {
		log.Printf("initial sign failed: %v", err)
	}
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if err := doSign(context.Background()); err != nil {
			log.Printf("sign failed: %v", err)
		}
	}

	// --- Strategy B: run once a day at 09:00 local time (default) ---
	// const runAtHour, runAtMinute = 9, 0

	// for {
	// 	time.Sleep(nextTick(runAtHour, runAtMinute))
	// 	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	// 	if err := doSign(ctx); err != nil {
	// 		log.Printf("sign failed: %v", err)
	// 	}
	// 	cancel()
	// }
}
