# Deploy terraform commands
# Setup Terraform from github actions @actions/terraform or something like that && Terraform Init && Terraform Plan && Terraform Apply

GO_FNS = oauth vault secret session_log audit_log
build_go: test_go
	@cd go && \
	$(foreach fn,$(GO_FNS), \
        GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -o ./$(fn)/bootstrap ./$(fn) || exit; \
        zip -j ../terraform/outputs/$(fn) $(fn)/bootstrap; \
        rm -f $(fn)/bootstrap; \
	)

test_go:
	@cd go && go test -cover ./... || exit;

NODE_FNS = session
build_node:
	@cd node && \
	$(foreach fn,$(NODE_FNS), \
		cd $(fn) && \
		npm install && npm test || exit && \
		zip -r -q ../../terraform/outputs/$(fn) . && cd ..; \
	)

mkdir_tf_outputs:
	@mkdir -p terraform/outputs
