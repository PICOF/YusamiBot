package hf

type SpaceConf struct {
	Components   []Component  `json:"components"`
	Dependencies []Dependency `json:"dependencies"`
	DevMode      bool         `json:"dev_mode"`
	EnableQueue  bool         `json:"enable_queue"`
	IsColab      bool         `json:"is_colab"`
	IsSpace      bool         `json:"is_space"`
	Layout       Layout       `json:"layout"`
	Mode         string       `json:"mode"`
	ShowAPI      bool         `json:"show_api"`
	ShowError    bool         `json:"show_error"`
	Theme        string       `json:"theme"`
	Title        string       `json:"title"`
	Version      string       `json:"version"`
}

type Func struct {
	Dependency *Dependency
	FnIndex    int
}

type Component struct {
	ID    int64 `json:"id"`
	Func  Func
	Props struct {
		Choices        []string    `json:"choices"`
		Components     []string    `json:"components"`
		ElemID         string      `json:"elem_id"`
		Headers        []string    `json:"headers"`
		Label          string      `json:"label"`
		Lines          int64       `json:"lines"`
		MaxLines       int64       `json:"max_lines"`
		Maximum        float64     `json:"maximum"`
		MinWidth       int64       `json:"min_width"`
		Minimum        float64     `json:"minimum"`
		Name           string      `json:"name"`
		Open           bool        `json:"open"`
		Samples        [][]string  `json:"samples"`
		SamplesPerPage int64       `json:"samples_per_page"`
		Scale          int64       `json:"scale"`
		ShowLabel      bool        `json:"show_label"`
		Source         string      `json:"source"`
		Step           float64     `json:"step"`
		Streaming      bool        `json:"streaming"`
		Style          struct{}    `json:"style"`
		Type           string      `json:"type"`
		Value          interface{} `json:"value"`
		Variant        string      `json:"variant"`
		Visible        bool        `json:"visible"`
	} `json:"props"`
	Type string `json:"type"`
}

type Dependency struct {
	APIName        interface{}   `json:"api_name"`
	BackendFn      bool          `json:"backend_fn"`
	Batch          bool          `json:"batch"`
	Cancels        []interface{} `json:"cancels"`
	Every          interface{}   `json:"every"`
	Inputs         []int64       `json:"inputs"`
	Js             string        `json:"js"`
	MaxBatchSize   int64         `json:"max_batch_size"`
	Outputs        []int64       `json:"outputs"`
	Queue          bool          `json:"queue"`
	ScrollToOutput bool          `json:"scroll_to_output"`
	ShowProgress   bool          `json:"show_progress"`
	Targets        []int64       `json:"targets"`
	Trigger        string        `json:"trigger"`
}

type Layout struct {
	ID       int64    `json:"id"`
	Children []Layout `json:"children"`
}
