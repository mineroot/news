package paging

import (
	"fmt"
	"strconv"
)

type Paginator struct {
	requestedPage int
	realPage      int
	size          int
	pagesCount    int
}

func NewPaginator(rawPage string, size int) (*Paginator, error) {
	if size < 1 {
		return nil, fmt.Errorf("size must be greater than zero")
	}

	if rawPage == "" {
		rawPage = "1"
	}
	requestedPage, err := strconv.Atoi(rawPage)
	if requestedPage <= 0 || err != nil {
		return nil, fmt.Errorf("invalid requested page: %w", err)
	}

	return &Paginator{
		requestedPage: requestedPage,
		realPage:      requestedPage,
		size:          size,
	}, nil
}

func (p *Paginator) SetRealItemsCount(itemsCount int) {
	if itemsCount == 0 {
		p.pagesCount = 1
		p.realPage = 1
		return
	}

	p.pagesCount = (itemsCount + p.size - 1) / p.size

	if p.realPage > p.pagesCount {
		p.realPage = p.pagesCount
	}
	if p.realPage < 1 {
		p.realPage = 1
	}
}

func (p *Paginator) NeedsRedirect() bool {
	return p.requestedPage != p.realPage
}

func (p *Paginator) Page() int {
	return p.realPage
}

func (p *Paginator) PagesCount() int {
	return p.pagesCount
}

func (p *Paginator) Size() int {
	return p.size
}
