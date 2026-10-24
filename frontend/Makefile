BUILD_DIR              = build
GULP                   = node_modules/.bin/gulp

devel:
	yarn install --frozen-lockfile
	$(GULP) devel

prod:
	yarn install --frozen-lockfile
	$(GULP) prod

clean:
	-rm -rf $(BUILD_DIR);
