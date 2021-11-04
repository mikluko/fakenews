package fakenews

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	hackernewsBaseURL            = "https://hacker-news.firebaseio.com/v0"
	hackernewsListPath           = "topstories.json"
	hackernewsItemPath           = "item"
	hackernewsDefaultConcurrency = 8
)

type hackernewsStoryList []int

type hackernewsStoryItem struct {
	Title string `json:"title"`
}

type HackernewsSource struct {
	Client      Client
	Limit       int
	Concurrency int
	items       []string
	mux         sync.Mutex
}

func (s *HackernewsSource) Fetch(ctx context.Context) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.items != nil {
		return nil
	}

	if s.Client == nil {
		s.Client = http.DefaultClient
	}
	if s.Concurrency == 0 {
		s.Concurrency = hackernewsDefaultConcurrency
	}

	list, err := s.fetchList(ctx)
	if err != nil {
		return err
	}
	if s.Limit > 0 {
		list = list[:s.Limit]
	}
	s.items = make([]string, len(list))

	g, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(s.Concurrency))
	for idx0, id0 := range list {
		err = sem.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		idx1, id1 := idx0, id0
		g.Go(func() (err error) {
			defer sem.Release(1)
			err = s.fetchItem(ctx, idx1, id1)
			if err != nil {
				return err
			}
			return nil
		})
	}
	return g.Wait()
}

func (s *HackernewsSource) Items() []string {
	return s.items
}

func (s *HackernewsSource) fetchItem(ctx context.Context, idx, id int) error {
	rq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/%d.json", hackernewsBaseURL, hackernewsItemPath, id), nil)
	if err != nil {
		return err
	}
	rs, err := s.Client.Do(rq.WithContext(ctx))
	if err != nil {
		return err
	}
	defer rs.Body.Close()
	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return err
	}
	var item hackernewsStoryItem
	err = json.Unmarshal(body, &item)
	if err != nil {
		return err
	}
	s.items[idx] = item.Title
	return nil
}

func (s *HackernewsSource) fetchList(ctx context.Context) (list hackernewsStoryList, err error) {
	rq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", hackernewsBaseURL, hackernewsListPath), nil)
	if err != nil {
		return
	}
	rs, err := s.Client.Do(rq.WithContext(ctx))
	if err != nil {
		return
	}
	defer rs.Body.Close()
	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &list)
	return
}
