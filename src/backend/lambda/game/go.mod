module georgguessr.com/lambda-game

go 1.15

require (
	georgguessr.com/pkg v0.0.0-00010101000000-000000000000
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go v1.38.42
)

replace georgguessr.com/pkg => ../../pkg
