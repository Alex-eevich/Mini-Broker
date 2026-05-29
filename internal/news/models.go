package news

type News struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Image       string `json:"image"`
	PublishedAt string `json:"published_at"`
	Categories  string `json:"categories"`
}

type FinnhubNews struct {
	Headline string `json:"headline"`
	Summary  string `json:"summary"`
	URL      string `json:"url"`
	Image    string `json:"image"`
	Datetime int64  `json:"datetime"`
	Source   string `json:"source"`
}
