build_service:
	go build ./cmd/places/main.go -c run_place_remember

build_indexer_scripts:
	go build ./cmd/indexer/main.go -c run_indexer_script

run_indexer_script:
	go run