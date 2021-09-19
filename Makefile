-include config.env
export
build_version := $(shell date +"%Y%m%d%H%M%S")

local:
	@echo "deploying locally";
	@cd frontend; ./make.sh $(build_version) "local";
	@cd backend; ./make.sh $(build_version) "local";
	@echo "$(build_version)";

remote:
	@echo "deploying remote";
	@cd backend; ./make.sh $(build_version) "remote";
	@cd frontend; ./make.sh $(build_version) "remote";
	@echo "$(build_version)";

guided:
	@echo "deploying remote";
	@cd backend; ./make.sh $(build_version) "remote" "guided";
	@cd frontend; ./make.sh $(build_version) "remote";
	@echo "$(build_version)";

remote_frontend:
	@echo "deploying remote frontend";
	@cd frontend; ./make.sh $(build_version) "remote";
	@echo "$(build_version)";

remote_backend:
	@echo "deploying remote backend";
	@cd backend; ./make.sh $(build_version) "remote";
	@echo "$(build_version)";s