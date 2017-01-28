Skip to content
This repository
Search
Pull requests
Issues
Gist
 @jmahowald
 Unwatch 5
  Star 0
 Fork 0 backpackhealth/batchpiper Private
 Code  Issues 0  Pull requests 0  Projects 0  Wiki  Pulse  Graphs  Settings
Branch: master Find file Copy pathbatchpiper/test/test_config.bats
650aa8c  on Dec 11, 2016
@deitch deitch Add tests for really long line to ensure yml does not break it
1 contributor
RawBlameHistory     
254 lines (226 sloc)  7.23 KB
# where will our test paths be?

load setup_tests

# now we can run our test cases
# we clear all outputs before each test
# to avoid questions about where it is mounting from, *everything*, including test data, is managed from a container
setup() {
	reset_test_data
}

######
#
# config vs environment vs options checks
#
######

# - config file undefined: default
@test "config file undefined" {
	EXPECTED="/batchpiper.yml"
	DUMP=$(docker run --rm $IMAGE --dump)
	VALUE=$(echo "$DUMP" | awk -F= '/^BATCHPIPE_CONFIG=/ {print $2}')
	[[ "$VALUE" == "$EXPECTED" ]]
}

# - config file in ENV: ENV
@test "config file in ENV" {
	EXPECTED="/q/r/f"
	DUMP=$(docker run -e BATCHPIPE_CONFIG=$EXPECTED --rm $IMAGE --dump)
	VALUE=$(echo "$DUMP" | awk -F= '/^BATCHPIPE_CONFIG=/ {print $2}')
	[[ "$VALUE" == "$EXPECTED" ]]
}

# - config file in CLI: CLI
@test "config file in CLI" {
	EXPECTED="/q/r/f"
	DUMP=$(docker run --rm $IMAGE --dump --config $EXPECTED)
	VALUE=$(echo "$DUMP" | awk -F= '/^BATCHPIPE_CONFIG=/ {print $2}')
	[[ "$VALUE" == "$EXPECTED" ]]
}

# - config file in CLI and ENV: CLI
@test "config file in CLI and ENV" {
	EXPECTED="/q/r/f"
	DUMP=$(docker run -e BATCHPIPE_CONFIG=/something/else --rm $IMAGE --dump --config $EXPECTED)
	VALUE=$(echo "$DUMP" | awk -F= '/^BATCHPIPE_CONFIG=/ {print $2}')
	[[ "$VALUE" == "$EXPECTED" ]]
}


check_env_only() {
	local envname=$1
	local envval=$2

	# - item in ENV: ENV
	DUMP=$(docker run -e "$envname"="$envval" --rm $IMAGE --dump)
	VALUE=$(echo "$DUMP" | awk -F= "/^$envname=/ "'{print $2}')
	[[ "$VALUE" == "$envval" ]]
}

check_config_only() {
	local envname=$1
	local fileval=$2

	# - item in file: file
	DUMP=$(docker run --rm -v $BASEDIR/config.yml:/batchpiper.yml $IMAGE --dump)
	VALUE=$(echo "$DUMP" | awk -F= "/^$envname=/ "'{print $2}')
	echo "DUMP VALUE fileval"
	echo "$DUMP"
	echo "$VALUE"
	echo "$fileval"
	[[ "$VALUE" == "$fileval" ]]
}
check_cli_only() {
	local envname=$1
	local optname=$2
	local optval=$3

	# - item in cli: cli
	DUMP=$(docker run --rm $IMAGE --$optname $optval --dump)
	VALUE=$(echo "$DUMP" | awk -F= "/^$envname=/ "'{print $2}')
	[[ "$VALUE" == "$optval" ]]
}
check_cli_and_env() {
	local envname=$1
	local envval=$2
	local optname=$3
	local optval=$4

	# - item in cli and env: cli
	DUMP=$(docker run --rm -e "$envname"="$envval" $IMAGE --$optname $optval --dump)
	VALUE=$(echo "$DUMP" | awk -F= "/^$envname=/ "'{print $2}')
	[[ "$VALUE" == "$optval" ]]
}
check_file_and_env() {
	local envname=$1
	local envval=$2
	local fileval=$3

	# - item in cli and env: cli
	DUMP=$(docker run --rm -e "$envname"="$envval" -v $BASEDIR/config.yml:/batchpiper.yml $IMAGE --dump)
	VALUE=$(echo "$DUMP" | awk -F= "/^$envname=/ "'{print $2}')
	[[ "$VALUE" == "$envval" ]]
}
check_cli_and_file() {
	local envname=$1
	local optname=$2
	local optval=$3
	local fileval=$4

	# - item in cli and env: cli
	DUMP=$(docker run --rm -v $BASEDIR/config.yml:/batchpiper.yml $IMAGE --$optname $optval --dump)
	VALUE=$(echo "$DUMP" | awk -F= "/^$envname=/ "'{print $2}')
	[[ "$VALUE" == "$optval" ]]
}
check_cli_and_file_and_env() {
	local envname=$1
	local envval=$2
	local optname=$3
	local optval=$4
	local fileval=$5

	# - item in cli and env: cli
	DUMP=$(docker run --rm -e "$envname"="$envval" -v $BASEDIR/config.yml:/batchpiper.yml $IMAGE --$optname $optval --dump)
	VALUE=$(echo "$DUMP" | awk -F= "/^$envname=/ "'{print $2}')
	[[ "$VALUE" == "$optval" ]]
}



# test how it processes configs

### THIS IS VERY UN-DRY... BUT bats HAS TERRIBLE SUPPORT FOR LOOPING AROUND @test

# output
@test "output env only" {
	check_env_only BATCHPIPE_OUTPUT /env/output
}
@test "output config file only" {
	check_config_only BATCHPIPE_OUTPUT /file/output
}
@test "output CLI only" {
	check_cli_only BATCHPIPE_OUTPUT output /opt/output
}
@test "output CLI and env var" {
	check_cli_and_env BATCHPIPE_OUTPUT /env/output output /opt/output
}
@test "output file and env var" {
	check_file_and_env BATCHPIPE_OUTPUT /env/output /file/output
}
@test "output CLI and file" {
	check_cli_and_file  BATCHPIPE_OUTPUT output /opt/output /file/output
}
@test "output CLI and file and env var" {
	check_cli_and_file_and_env BATCHPIPE_OUTPUT /env/output output /opt/output /file/output
}

# pluginBase
@test "pluginBase env only" {
	check_env_only BATCHPIPE_PLUGINBASE /env/base
}
@test "pluginBase config file only" {
	check_config_only BATCHPIPE_PLUGINBASE /file/base
}
@test "pluginBase CLI only" {
	check_cli_only BATCHPIPE_PLUGINBASE pluginBase /opt/base
}
@test "pluginBase CLI and env var" {
	check_cli_and_env BATCHPIPE_PLUGINBASE /env/base pluginBase /opt/base
}
@test "pluginBase file and env var" {
	check_file_and_env BATCHPIPE_PLUGINBASE /env/base /file/base
}
@test "pluginBase CLI and file" {
	check_cli_and_file  BATCHPIPE_PLUGINBASE pluginBase /opt/base /file/base
}
@test "pluginBase CLI and file and env var" {
	check_cli_and_file_and_env BATCHPIPE_PLUGINBASE /env/base pluginBase /opt/base /file/base
}

# batch
@test "batch env only" {
	check_env_only BATCHPIPE_BATCH batch@env
}
@test "batch config file only" {
	check_config_only BATCHPIPE_BATCH "batch1 file1,batch2 file2,batch3isreallylong andithasonereallylongsourceyouknow andithasanotherreallylongsource totestlinebreaks"
}
@test "batch CLI only" {
	check_cli_only BATCHPIPE_BATCH batch batch@opt
}
@test "batch CLI and env var" {
	check_cli_and_env BATCHPIPE_BATCH batch@env batch batch@opt
}
@test "batch file and env var" {
	check_file_and_env BATCHPIPE_BATCH batch@env batch@opt
}
@test "batch CLI and file" {
	check_cli_and_file  BATCHPIPE_BATCH batch batch@opt batch@file
}
@test "batch CLI and file and env var" {
	check_cli_and_file_and_env BATCHPIPE_BATCH batch@env batch batch@opt batch@file
}

# pipeline
@test "pipeline env only" {
	check_env_only BATCHPIPE_PIPELINE pipelineEnv
}
@test "pipeline config file only" {
	check_config_only BATCHPIPE_PIPELINE "pipeline1 file1,pipeline2"
}
@test "pipeline CLI only" {
	check_cli_only BATCHPIPE_PIPELINE pipeline pipelineOpt
}
@test "pipeline CLI and env var" {
	check_cli_and_env BATCHPIPE_PIPELINE pipelineEnv pipeline pipelineOpt
}
@test "pipeline file and env var" {
	check_file_and_env BATCHPIPE_PIPELINE pipelineEnv pipelineFile
}
@test "pipeline CLI and file" {
	check_cli_and_file  BATCHPIPE_PIPELINE pipeline pipelineOpt pipelineFile
}
@test "pipeline CLI and file and env var" {
	check_cli_and_file_and_env BATCHPIPE_PIPELINE pipelineEnv pipeline pipelineOpt pipelineFile
}


# custom args
@test "custom not defined" {
	DUMP=$(docker run --rm $IMAGE --dump)
	VALUE=$(echo "$DUMP" | awk -F= "/^BATCHPIPE_CUSTOMARGS=/ "'{print $2}')
	[[ -z "$VALUE" ]]
}
@test "custom file only" {
	DUMP=$(docker run --rm -v $BASEDIR/config.yml:/batchpiper.yml $IMAGE --dump)
	VALUE=$(echo "$DUMP" | awk -F= "/^BATCHPIPE_CUSTOMARGS=/ "'{print $2}')
	echo "DUMP VALUE"
	echo "$DUMP"
	echo "$VALUE"
	[[ "$VALUE" == "fileCustomA fileCustomB" ]]
}
@test "custom CLI only" {
	local optVal="--opCustomA valCustomA --optCustomB"
	DUMP=$(docker run --rm $IMAGE --dump -- $optVal)
	VALUE=$(echo "$DUMP" | awk -F= "/^BATCHPIPE_CUSTOMARGS=/ "'{print $2}')
	[[ "$VALUE" == "$optVal" ]]
}
@test "custom file and CLI" {
	local optVal="--opCustomA valCustomA --optCustomB"
	DUMP=$(docker run --rm -v $BASEDIR/config.yml:/batchpiper.yml $IMAGE --dump -- $optVal)
	VALUE=$(echo "$DUMP" | awk -F= "/^BATCHPIPE_CUSTOMARGS=/ "'{print $2}')
	[[ "$VALUE" == "$optVal" ]]
}
Contact GitHub API Training Shop Blog About
© 2017 GitHub, Inc. Terms Privacy Security Status Help
