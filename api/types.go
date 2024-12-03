package api

// AddSongRequest описывает структуру входящего JSON для добавления песни.
type AddSongRequest struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

// SongDetail описывает структуру данных ответа на запрос деталей песни.
type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}
