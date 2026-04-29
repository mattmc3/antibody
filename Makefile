# Delegate all targets to justfile
.DEFAULT_GOAL := build

%:
	@just $@

.PHONY: setup test cover lint ci build fmt
