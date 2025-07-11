.PHONY: default 
default: check

.PHONY: check 
check:
	go tool lefthook run pre-commit
