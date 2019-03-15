package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

func isFileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func getPhotosURLs(pages int) []string {
	c := colly.NewCollector()
	c.Async = false
	c.UserAgent = "ImagesCrawler"
	c.AllowedDomains = append(c.AllowedDomains, "trainpix.org")

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       time.Second,
		RandomDelay: time.Second,
	})
	if err != nil {
		log.Fatalf("Error while setting Colly limits: %v", err)
	}

	result := []string{}

	re, err := regexp.Compile("/(\\d+)_s\\.jpg$")
	if err != nil {
		log.Fatalf("Can't compile RexExp: %v", err)
	}

	c.OnHTML(".main img.f", func(e *colly.HTMLElement) {
		smallImagePath := e.Attr("src")
		if smallImagePath == "" {
			log.Printf("Warning: some image have empty image path")
			return
		}
		fullImagePath := re.ReplaceAllString(smallImagePath, "/${1}.jpg")
		if !strings.HasPrefix(strings.ToLower(fullImagePath), "http") {
			fullImagePath = "https://trainpix.org" + fullImagePath
		}
		result = append(result, fullImagePath)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	c.OnScraped(func(r *colly.Response) {
		log.Printf("Scraped %s\n", r.Request.URL.String())
	})

	for pg := 0; pg < pages; pg++ {
		url := fmt.Sprintf(
			"https://trainpix.org/voting.php?show=results&st=%d",
			pg*10,
		)

		countBefore := len(result)
		c.Visit(url)
		countAfter := len(result)
		if countBefore == countAfter {
			break
		}
		log.Printf("Fetched %d pictures URLs", countAfter-countBefore)
	}

	return result
}

// If the photo is alrady downloaded the function will return true orherwise false
func downloadPhoto(
	url, dir, upscalesDir string,
	allowHorizontal, allowVertical bool,
) (bool, error) {
	re, err := regexp.Compile("/(\\d+)\\.jpg$")
	if err != nil {
		log.Fatalf("Can't compile RexExp: %v", err)
	}
	match := re.FindStringSubmatch(url)
	if match == nil {
		return false, fmt.Errorf("can't extract photo ID from '%v'", url)
	}
	photoID := match[1]

	photoPath := filepath.Join(dir, fmt.Sprintf("%s.jpg", photoID))
	{
		exists, err := isFileExist(photoPath)
		if err != nil {
			return false, fmt.Errorf("can't look to file: %v", err)
		}
		if exists {
			log.Printf("Photo %s already downloaded", photoID)
			return true, nil
		}

		if upscalesDir != "" {
			upscaledPhotoPath := filepath.Join(upscalesDir, fmt.Sprintf("%s.jp2", photoID))
			upscaledExists, _ := isFileExist(upscaledPhotoPath)
			if upscaledExists {
				log.Printf("Upscaled photo %s already present", photoID)
				return true, nil
			}
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("can't do request: %v", err)
	}
	req.Header.Set("User-Agent", "ImagesCrawler")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("can't do request: %v", err)
	}
	defer resp.Body.Close()

	tempPhotoPath := filepath.Join(dir, fmt.Sprintf("%s.tmp.jpg", photoID))
	{
		exist, err := isFileExist(tempPhotoPath)
		if err != nil {
			return false, fmt.Errorf("can't look to temp file: %v", err)
		}
		if exist {
			err := os.Remove(tempPhotoPath)
			if err != nil {
				return false, fmt.Errorf("can't remove temp file: %v", err)
			}
		}
	}

	file, err := os.Create(tempPhotoPath)
	if err != nil {
		return false, fmt.Errorf("unable to create a file: %v", err)
	}
	defer file.Close()
	defer os.Remove(tempPhotoPath)

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return false, fmt.Errorf("unable to write a photo: %v", err)
	}
	err = file.Close()
	if err != nil {
		return false, fmt.Errorf("unable to close photo file: %v", err)
	}

	err = os.Rename(tempPhotoPath, photoPath)
	if err != nil {
		return false, fmt.Errorf("can't rename photo file: %v", err)
	}

	return false, nil
}

func main() {
	pages := flag.Int(
		"pages",
		10,
		"\"-pages 10\" will take pictures from 10 first pages,"+
			" if \"pages\" will be -1 then crawler will take all pages",
	)
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dir := flag.String(
		"dir",
		cwd,
		"where we should save pictures",
	)
	upscalesDir := flag.String(
		"upscales",
		"",
		"directory where upscaled files stored. With this option you can "+
			"delete downloaded images keeping only upscales",
	)
	allowHorizontal := flag.Bool(
		"horiz",
		true,
		"allow download horizontal pictures? [-horiz=false], [-horiz=true]",
	)
	allowVertical := flag.Bool(
		"vert",
		false,
		"allow download vertical pictures? [-vert=false], [-vert=true]",
	)
	flag.Parse()
	if *allowHorizontal == false && *allowVertical == false {
		log.Fatal("Can't start: vertical and horizontal photos are disallowed")
	}

	log.Print("Started")
	log.Printf("Using pages = %d, dir = %s", *pages, *dir)
	if *upscalesDir != "" {
		log.Printf("Using upscales dir = %s", *upscalesDir)
	} else {
		log.Printf("Working without upscales dir")
	}
	log.Printf(
		"<allow horizontal> = %v, <allow vertical> = %v",
		*allowHorizontal, *allowVertical,
	)
	if *pages < 0 {
		*pages = math.MaxUint32 >> 1
	}

	photosURLs := getPhotosURLs(*pages)
	for _, photoURL := range photosURLs {
		log.Printf("Downloading %s", photoURL)
		isAlreadyDownloaded, err := downloadPhoto(
			photoURL, *dir, *upscalesDir,
			*allowHorizontal, *allowVertical,
		)
		if err != nil {
			log.Printf(
				"[Warning] Error while donloading a photo %s: %v",
				photoURL,
				err,
			)
		}
		if !isAlreadyDownloaded {
			time.Sleep(time.Second)
		}
	}
	log.Println("All done")
}
