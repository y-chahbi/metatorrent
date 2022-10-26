package providerCpasbien

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/antchfx/htmlquery"
	. "github.com/demostanis/metatorrent/internal/messages"
)

const Name = "cpasbien"
const MainUrl = "https://www.cpasbien.ch"

var providerCpasienError = func(category, msg string) error {
	return errors.New(fmt.Sprintf("[%s/%s] ERROR: %s", Name, category, msg))
}

func searchPage(query string, beginning int, end int, statusChannel chan StatusMsg, torrentsChannel chan TorrentsMsg) error {
	var wg sync.WaitGroup
	doc, err := htmlquery.LoadURL(fmt.Sprintf("%s/recherche/%s/%d", MainUrl, query, beginning))
	if err != nil {
		return err
	}

	status(statusChannel, "[%s] Processing %dth torrents...", Name, beginning)

	titleElements, _ := htmlquery.QueryAll(doc, "//td/a[@class=\"titre\"]/text()")
	linkElements, _ := htmlquery.QueryAll(doc, "//td/a[@class=\"titre\"]/@href")
	sizeElements, _ := htmlquery.QueryAll(doc, "//td/div[@class=\"poid\"]")
	seedersCountElements, _ := htmlquery.QueryAll(doc, "//td/div[@class=\"down\"]")
	leechersCountElements, _ := htmlquery.QueryAll(doc, "//td/div[@class=\"up\"]//text()")

	if len(linkElements) != len(titleElements) ||
		len(seedersCountElements) != len(titleElements) ||
		len(leechersCountElements) != len(titleElements) ||
		len(sizeElements) != len(titleElements) {
		return providerCpasienError("parsing", "Torrent entries are malformed.")
	}

	for i, title := range titleElements {
		title := htmlquery.InnerText(title)
		link := htmlquery.SelectAttr(linkElements[i], "href")

		seeders, err := strconv.Atoi(htmlquery.InnerText(seedersCountElements[i]))
		if err != nil {
			return providerCpasienError("parsing", "Expected seeders to be a number.")
		}
		leechers, err := strconv.Atoi(htmlquery.InnerText(leechersCountElements[i]))
		if err != nil {
			return providerCpasienError("parsing", "Expected leechers to be a number.")
		}
		size := htmlquery.InnerText(sizeElements[i])

		myTorrent := ProviderCpasbienTorrent{
			title:    title,
			link:     link,
			seeders:  seeders,
			leechers: leechers,
			size:     size,
		}
		wg.Add(1)
		go func() {
			torrentsChannel <- myTorrent
			wg.Done()
		}()
	}

	status(statusChannel, "[%s] Processed %d torrents...", Name, len(titleElements))
	wg.Wait()
	return nil
}

func Search(query string, statusChannel chan StatusMsg, torrentsChannel chan TorrentsMsg, errorsChannel chan ErrorsMsg) {
	doc, err := htmlquery.LoadURL(fmt.Sprintf("%s/recherche/%s", MainUrl, query))
	if err != nil {
		errorsChannel <- err
		return
	}

	pages := htmlquery.Find(doc, "//ul[@class=\"pagination\"]//a")
	if len(pages) == 0 {
		errorsChannel <- providerCpasienError("parsing", "Number of torrents missing.")
		return
	}
	indexes := make([]int, 0)
	for _, page := range pages {
		matches := regexp.MustCompile(`\[(\d+)-\d+\]`).FindStringSubmatch(htmlquery.InnerText(page))
		if len(matches) > 0 {
			beginning, _ := strconv.Atoi(matches[1])
			indexes = append(indexes, beginning)
		}
	}
	lastIndex := indexes[len(indexes)-1]
	pageCount := len(indexes)

	status(statusChannel, "[%s] Found %d pages", Name, pageCount)

	var lastError error
	var wg sync.WaitGroup

	for i := 0; i < pageCount; i++ {
		wg.Add(1)
		go func(i int) {
			err := searchPage(query, indexes[i], lastIndex, statusChannel, torrentsChannel)
			wg.Done()
			if err != nil {
				lastError = err
				return
			}
		}(i)
	}

	wg.Wait()
	finalStatus(statusChannel, "[%s] Done", Name)
	if lastError != nil {
		errorsChannel <- lastError
		return
	}
}

func status(statusChannel chan StatusMsg, message string, rest ...any) {
	go func() {
		statusChannel <- StatusMsg{fmt.Sprintf(message, rest...), false}
	}()
}

func finalStatus(statusChannel chan StatusMsg, message string, rest ...any) {
	go func() {
		statusChannel <- StatusMsg{fmt.Sprintf(message, rest...), true}
	}()
}