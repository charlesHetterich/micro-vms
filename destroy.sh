HOST="0.0.0.0:9090"

for uid in $(fl microvm get --host "$HOST" 2>&1 \
             | awk '/^[0-9A-Z]{26}/ {print $1}'); do
  echo "deleting $uid"
  fl microvm delete --host "$HOST" "$uid"
done

# sudo journalctl -u flintlockd -n 50 -f
# sudo journalctl -u flintlockd | grep "$UID" -n -B2 -A4 | less

# sudo journalctl -u flintlockd | grep 01JZXKE4NNVTTRY9E0G58ZQX1F -n -B2 -A4 | less

# fl microvm get --host 0.0.0.0:9090