module georgguessr.com/lambda-room

go 1.15

replace georgguessr.com/pkg => ../../pkg

require (
	georgguessr.com/pkg v0.0.0-00010101000000-000000000000
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go v1.38.42
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/mmcloughlin/globe v0.0.0-20200201185603-653bb586373c // indirect
	github.com/mmcloughlin/spherand v0.0.0-20200201191112-cd5c4c9261aa
	github.com/paulmach/orb v0.2.1
	github.com/tidwall/pinhole v0.0.0-20210130162507-d8644a7c3d19 // indirect
	golang.org/x/image v0.0.0-20210504121937-7319ad40d33e // indirect
)
