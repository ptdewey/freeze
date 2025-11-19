clean-test:
    @rm -rf ./__snapshots__
    @go test ./... -cover -coverprofile=cover.out

test:
    @go test ./... -cover -coverprofile=cover.out

run:
    @go run cmd/freeze/main.go

clean:
    @rm -rf ./__snapshots__

tui:
    @pushd ./cmd/tui && go build -o freeze-tui ./main.go && popd
    @./cmd/tui/freeze-tui

review:
    @./cmd/tui/freeze-tui
