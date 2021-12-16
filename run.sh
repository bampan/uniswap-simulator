#!/bin/bash

#SBATCH --mail-type=ALL                           # mail configuration: NONE, BEGIN, END, FAIL, REQUEUE, ALL
#SBATCH --output=/home/yhuynh/uniswap-simulatorlog/%j.out     # where to store the output (%j is the JOBID), subdirectory "log" must exist
#SBATCH --error=/home/yhuynh/uniswap-simulator/log/%j.err  # where to store error messages
#SBATCH --cpus-per-task=40
#SBATCH --mem=32G
# Exit on errors
set -o errexit

# Send some noteworthy information to the output log
echo "Running on node: $(hostname)"
echo "In directory:    $(pwd)"
echo "Starting on:     $(date)"
echo "SLURM_JOB_ID:    ${SLURM_JOB_ID}"

# Binary or script to execute
/home/yhuynh/uniswap-simulator/main

# Send more noteworthy information to the output log
echo "Finished at:     $(date)"

# End the script with exit code 0
exit 0