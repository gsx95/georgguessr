-include config.env
export

local:
	@echo "deploying locally";
	@cd frontend; ./make.sh "local";
	@cd backend; ./make.sh "local";

remote:
	@echo "deploying remote";
	@cd backend; ./make.sh "remote";
	@cd frontend; ./make.sh "remote";

guided:
	@echo "deploying remote";
	@cd backend; ./make.sh "remote" "guided";
	@cd frontend; ./make.sh "remote";

remote_frontend:
	@echo "deploying remote frontend";
	@cd frontend; ./make.sh "remote";

remote_backend:
	@echo "deploying remote backend";
	@cd backend; ./make.sh "remote";