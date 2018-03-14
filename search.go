package search

import "github.com/bitly/go-simplejson"
import "io/ioutil"
import "net/http"
import "net/url"
import "strconv"
import "fmt"
import "strings"

// Loggly search client with user credentials, loggly
// does not seem to support tokens right now.
type Client struct {
	User         string
	Pass         string
	Account      string
	Endpoint     string
	Paginating   bool
	GetAllAtOnce bool
}

// Search response with total events, page number
// and the events array.
type Response struct {
	Total  int64
	Page   int64
	Events []interface{}
	Url    string
}

// Query builder struct
type query struct {
	client *Client
	query  string
	from   string
	until  string
	order  string
	size   int
}

// Create a new query
func newQuery(c *Client, str string, ord string) *query {
	return &query{
		client: c,
		query:  str,
		from:   "-24h",
		until:  "now",
		order:  ord,
		size:   100,
	}
}

// Create a new loggly search client with credentials.
func New(account string, user string, pass string) *Client {
	c := &Client{
		Account:      account,
		User:         user,
		Pass:         pass,
		Endpoint:     "loggly.com/apiv2",
		Paginating:   true,
		GetAllAtOnce: false,
	}

	return c
}

// Return the base api url.
func (c *Client) Url() string {
	return fmt.Sprintf("https://%s:%s@%s.%s", c.User, c.Pass, c.Account, c.Endpoint)
}

// GET the given path.
func (c *Client) Get(path string) (*http.Response, error) {
	return http.Get(c.Url() + path)
}

// GET json from the given path.
func (c *Client) GetJSON(path string) (j *simplejson.Json, err error) {
	res, err := c.Get(path)

	if err != nil {
		return
	}

	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("go-loggly-search: %q", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}

	j, err = simplejson.NewJson(body)
	return
}

// Create a new search instance, loggly requires that a search
// is made before you may fetch events from it with a second call.
func (c *Client) CreateSearch(params string) (*simplejson.Json, error) {
	if c.Paginating {
		return c.GetJSON("/events/iterate?" + params)
	} else {
		return c.GetJSON("/search?" + params)
	}
}

// Create a next paginating search instance.
func (c *Client) CreateNextSearch(params string) (*simplejson.Json, error) {
	return c.GetJSON("/events/iterate?next=" + params)
}

// Get events, must be called after CreateSearch() with the
// correct rsid to reference the search.
func (c *Client) GetEvents(params string) (*simplejson.Json, error) {
	return c.GetJSON("/events?" + params)
}

// Search response with total events, page number
// and the events array.
func (c *Client) Search(params string) (*Response, error) {
	j, err := c.CreateSearch(params)

	if err != nil {
		return nil, err
	}

	urlNext := ""

	if c.Paginating {
		urlNext = j.Get("next").MustString()
	} else {
		id := j.GetPath("rsid", "id").MustString()
		j, err = c.GetEvents("rsid=" + id)

		if err != nil {
			return nil, err
		}
	}

	// Search response with total events, page number
	// and the events array.
	if c.Paginating {
		response := &Response{
			Events: j.Get("events").MustArray(),
			Url:    urlNext,
		}
		if c.GetAllAtOnce {
			for response.Url != "" {
				s := strings.Split(response.Url, "=")
				nextID := s[1]

				j, err = c.CreateNextSearch(nextID)
				if err != nil {
					return nil, err
				}
				response.Url = j.Get("next").MustString()
				response.Events = append(response.Events, j.Get("events").MustArray()...)
			}
		}
		return response, nil
	} else {
		return &Response{
			Total:  j.Get("total_events").MustInt64(),
			Page:   j.Get("page").MustInt64(),
			Events: j.Get("events").MustArray(),
		}, nil
	}
}

// NextSearch response with next part of events
// for the paginating mode on. It also may
// contains url for the next part of events
func (c *Client) NextSearch(nextID string) (*Response, error) {
	j, err := c.CreateNextSearch(nextID)
	if err != nil {
		return nil, err
	}
	return &Response{
		Events: j.Get("events").MustArray(),
		Url:    j.Get("next").MustString(),
	}, nil
}

// Create a new search query using the fluent api.
func (c *Client) Query(str string, order string) *query {
	return newQuery(c, str, order)
}

// Return the encoded query-string.
func (q *query) String() string {
	qs := url.Values{}
	qs.Set("q", q.query)
	qs.Set("size", strconv.Itoa(q.size))
	qs.Set("from", q.from)
	qs.Set("until", q.until)
	qs.Set("order", q.order)
	return qs.Encode()
}

// Set response size.
func (q *query) Size(n int) *query {
	q.size = n
	return q
}

// Set from time.
func (q *query) From(str string) *query {
	q.from = str
	return q
}

// Set until time.
func (q *query) Until(str string) *query {
	q.until = str
	return q
}

// Set until time.
func (q *query) To(str string) *query {
	q.until = str
	return q
}

// Fetch response with events, page number
// and the events array. In Paginating mode
// it also may contains url for the next
// part of events
func (q *query) Fetch() (*Response, error) {
	return q.client.Search(q.String())
}

// NextFetch response with next part of events
// for the paginating mode on. It also may
// contains url for the next part of events
func (q *query) NextFetch(nextID string) (*Response, error) {
	return q.client.NextSearch(nextID)
}
