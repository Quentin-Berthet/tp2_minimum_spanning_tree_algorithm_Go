BIN_FOLDER=./bin

all: tp2_minimum_spanning_tree

tp2_minimum_spanning_tree: main.go neighbour
	go build -o $(BIN_FOLDER)/$@

.PHONY: clean
clean:
	rm -f $(BIN_FOLDER)/tp2_minimum_spanning_tree
