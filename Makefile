# Makefile for go app
# author: julien-beguier
GSSERVER=	fizzbuzz-server
BUILD_DIR=	bin
GOBUILD=	go build
RM=		rm -f

all: server

server:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o ./$(BUILD_DIR)/$(GSSERVER) main.go
	@echo "\n\033[1;31mBuild $(GSSERVER) complete\033[0;0m\n"

fclean:
	$(RM) $(BIN)/$(GSSERVER)
	@echo "\n\033[1;31mRemoved \033[1;33m$(BIN)/$(GSSERVER)\033[0;0m\n"

re:	fclean server

.PHONY:	all server fclean re
