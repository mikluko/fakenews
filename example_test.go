package fakenews_test

import (
	"context"
	"log"

	"github.com/mikluko/fakenews"
)

func Example() {
	fn := fakenews.NewGenerator(&fakenews.HackernewsSource{Limit: 100, Concurrency: 4})
	if err := fn.Init(context.TODO()); err != nil {
		log.Fatalln(err)
	}
	for i := 0; i < 10; i++ {
		item, err := fn.Generate()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(item)
	}
	// Output:
}
