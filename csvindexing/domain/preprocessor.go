package domain

type (
	// ProductRowPreprocessor can edit and validate incoming product rows
	ProductRowPreprocessor interface {
		Preprocess(row map[string]string, options ProductRowPreprocessOptions) (map[string]string, error)
	}

	// ProductRowPreprocessOptions for the Preprocessor
	ProductRowPreprocessOptions struct {
		Locale   string
		Currency string
	}

	// CategoryRowPreprocessor can edit and validate incoming category rows
	CategoryRowPreprocessor interface {
		Preprocess(row map[string]string, options CategoryRowPreprocessOptions) (map[string]string, error)
	}

	// CategoryRowPreprocessOptions for the Preprocessor
	CategoryRowPreprocessOptions struct {
		Locale string
	}
)
