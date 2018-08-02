#!/usr/bin/env bash

TEST_SCRIPT=$(cat test_script_pipeline_template)
printf "############### Running test pipeline with antlr script parser ###############\n\n"
go build -o bitflow-antlr github.com/antongulenko/go-bitflow-pipeline/bitflowcli/cmd
./bitflow-antlr "${TEST_SCRIPT} test_out_antlr.out"
rm ./bitflow-antlr

printf "\n############### Running test pipeline with original script parser ###############\n\n"
go build -o bitflow-original github.com/antongulenko/go-bitflow-pipeline/bitflow-pipeline
./bitflow-original "${TEST_SCRIPT}test_out_original.out"
rm ./bitflow-original

rm test_out*.out