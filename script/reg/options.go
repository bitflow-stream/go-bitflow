package reg

type Options struct {
	RequiredParams []string
	OptionalParams []string

	// SupportBatchProcessing, if true this Processor can be called with batches
	SupportBatchProcessing bool

	// EnforceBatchProcessing, if true this Processor can ONLY be called with batches
	EnforceBatchProcessing bool
}

type Option func(*Options)

func RequiredParams(params ...string) Option {
	return func(opts *Options) {
		opts.RequiredParams = append(opts.RequiredParams, params...)
	}
}

func OptionalParams(params ...string) Option {
	return func(opts *Options) {
		opts.OptionalParams = append(opts.OptionalParams, params...)
	}
}

func SupportBatch() Option {
	return func(opts *Options) {
		opts.SupportBatchProcessing = true
	}
}

func EnforceBatch() Option {
	return func(opts *Options) {
		opts.SupportBatchProcessing = true
		opts.EnforceBatchProcessing = true
	}
}

func GetOpts(options []Option) Options {
	opts := Options{}
	for _, o := range options {
		o(&opts)
	}
	return opts
}
