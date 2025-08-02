HOST="0.0.0.0:9090"
COUNT=$(fl microvm get --host "$HOST" 2>&1 \
             | awk '/^[0-9A-Z]{26}/ {print $1}' | wc -l)
echo $((COUNT + 10))
