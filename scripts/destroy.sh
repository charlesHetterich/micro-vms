HOST="0.0.0.0:9090"

for uid in $(fl microvm get --host "$HOST" 2>&1 \
             | awk '/^[0-9A-Z]{26}/ {print $1}'); do
  echo "deleting $uid"
  fl microvm delete --host "$HOST" "$uid"
done
rm ~/.ssh/known_hosts