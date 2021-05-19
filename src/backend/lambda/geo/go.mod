module georgguessr.com/lambda-geo

go 1.15

replace georgguessr.com/pkg => ../../pkg

require (
	georgguessr.com/pkg v0.0.0-00010101000000-000000000000
	github.com/aws/aws-lambda-go v1.23.0
)
