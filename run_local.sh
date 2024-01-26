experiment_duration=$1
builder=$2
validator=$3
regular=$4
parcel_size=$5
log_output_dir=$6

# ========== Experiment Launch ==========
echo "========== Experiment Launch =========="

# Run validator
if [ "$validator" -ne 0 ]; then
    for ((i=0; i<$validator; i++)); do
        go run . -duration="$experiment_duration" -nodeType=validator -size="$parcel_size" -logOutput="$log_output_dir" &
        echo "validator $i"
        sleep 0.5
    done

    if [ "$builder" -eq 0 ] && [ "$regular" -ne 0 ]; then
        go run . -duration"=$experiment_duration" -nodeType=validator -size="$parcel_size" -logOutput="$log_output_dir" 
    else
        if [ "$validator" -ne 1 ]; then
            go run . -duration="$experiment_duration" -nodeType=validator -size="$parcel_size" -logOutput="$log_output_dir"&
            sleep 0.5
        fi
    fi
fi

# Run other nodes
if [ "$regular" -ne 0 ]; then
    for ((i=0; i<$regular; i++)); do
        go run . -duration="$experiment_duration" -nodeType=regular -size="$parcel_size" -logOutput="$log_output_dir" &
        echo "regular $i"
        sleep 0.5
    done

    if [ "$builder" -eq 0 ]; then
        go run . -duration="$experiment_duration" -nodeType=regular -size="$parcel_size" -logOutput="$log_output_dir"
    else
        if [ "$regular" -ne 1 ]; then
            go run . -duration="$experiment_duration" -nodeType=regular -size="$parcel_size" -logOutput="$log_output_dir" &
            sleep 0.5
        fi
    fi

fi

if [ "$builder" -ne 0 ]; then
    echo "builder launch"
    go run . -duration="$experiment_duration" -nodeType=builder -size="$parcel_size" -logOutput="$log_output_dir"
fi