package s3

import "go.uber.org/fx"

func NewS3ClientFx() fx.Option {
	return fx.Module("s3",
		fx.Provide(NewAWSConfig, NewClient, NewService),
	)
}
