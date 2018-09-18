#!/usr/bin/env bash

TEST_SCRIPT=$(cat test_script_pipeline_template)
go build -o bitflow-test github.com/antongulenko/go-bitflow-pipeline/bitflow-pipeline

printf "\n############### Running test pipeline with original script parser ###############\n\n"
./bitflow-test "${TEST_SCRIPT}test_out_original.out"

printf "############### Running test pipeline with antlr script parser ###############\n\n"
./bitflow-test -new "${TEST_SCRIPT} test_out_antlr.out"

rm test_out*.out
rm ./bitflow-test