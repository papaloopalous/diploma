ENV_FILE := .env
API_CONFIG_OUTPUT := ./configs/api.yaml
API_CONFIG_TEMPLATE := api_template.txt
BALANCER_CONFIG_OUTPUT := ./configs/load_balancer.yaml
BALANCER_CONFIG_TEMPLATE := balancer_template.txt
LOGGER_CONFIG_OUTPUT := ./configs/logger.env
LOGGER_CONFIG_TEMPLATE := logger_template.txt
POSTGRE_CONFIG_OUTPUT := ./configs/postgre_api.env
POSTGRE_CONFIG_TEMPLATE := postgre_template.txt
TARANTOOL_CONFIG_OUTPUT := ./configs/tarantool_api.env
TARANTOOL_CONFIG_TEMPLATE := tarantool_template.txt
PROMETHEUS_CONFIG_OUTPUT := ./configs/prometheus.yaml
PROMETHEUS_CONFIG_TEMPLATE := prometheus_template.txt


.PHONY: all generate_api generate_balancer generate_logger generate_postgre generate_tarantool generate_prometheus

generate_api:
	@echo "Generating $(API_CONFIG_OUTPUT) from $(API_CONFIG_TEMPLATE)..."
	@export $$(cat $(ENV_FILE) | sed 's/ *= */=/' | grep -v '^#') && \
	envsubst < $(API_CONFIG_TEMPLATE) > $(API_CONFIG_OUTPUT)
	@echo "Done: $(API_CONFIG_OUTPUT) created."

generate_balancer:
	@echo "Generating $(BALANCER_CONFIG_OUTPUT) from $(BALANCER_CONFIG_TEMPLATE)..."
	@export $$(cat $(ENV_FILE) | sed 's/ *= */=/' | grep -v '^#') && \
	envsubst < $(BALANCER_CONFIG_TEMPLATE) > $(BALANCER_CONFIG_OUTPUT)
	@echo "Done: $(BALANCER_CONFIG_OUTPUT) created."

generate_logger:
	@echo "Generating $(LOGGER_CONFIG_OUTPUT) from $(LOGGER_CONFIG_TEMPLATE)..."
	@export $$(cat $(ENV_FILE) | sed 's/ *= */=/' | grep -v '^#') && \
	envsubst < $(LOGGER_CONFIG_TEMPLATE) > $(LOGGER_CONFIG_OUTPUT)
	@echo "Done: $(LOGGER_CONFIG_OUTPUT) created."

generate_postgre:
	@echo "Generating $(POSTGRE_CONFIG_OUTPUT) from $(POSTGRE_CONFIG_TEMPLATE)..."
	@export $$(cat $(ENV_FILE) | sed 's/ *= */=/' | grep -v '^#') && \
	envsubst < $(POSTGRE_CONFIG_TEMPLATE) > $(POSTGRE_CONFIG_OUTPUT)
	@echo "Done: $(POSTGRE_CONFIG_OUTPUT) created."

generate_tarantool:
	@echo "Generating $(TARANTOOL_CONFIG_OUTPUT) from $(TARANTOOL_CONFIG_TEMPLATE)..."
	@export $$(cat $(ENV_FILE) | sed 's/ *= */=/' | grep -v '^#') && \
	envsubst < $(TARANTOOL_CONFIG_TEMPLATE) > $(TARANTOOL_CONFIG_OUTPUT)
	@echo "Done: $(TARANTOOL_CONFIG_OUTPUT) created."

generate_prometheus:
	@echo "Generating $(PROMETHEUS_CONFIG_OUTPUT) from $(PROMETHEUS_CONFIG_TEMPLATE)..."
	@export $$(cat $(ENV_FILE) | sed 's/ *= */=/' | grep -v '^#') && \
	envsubst < $(PROMETHEUS_CONFIG_TEMPLATE) > $(PROMETHEUS_CONFIG_OUTPUT)
	@echo "Done: $(PROMETHEUS_CONFIG_OUTPUT) created."

all: generate_api generate_balancer generate_logger generate_postgre generate_tarantool generate_prometheus