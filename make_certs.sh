rm .cert.pem 2>/dev/null
rm .cert.key 2>/dev/null
echo "creating server cert"
# customize this command with server hostname/DNS
openssl req -x509  -sha256 -days 3650 -nodes -keyout .cert.key -out .cert.pem -extensions san -config <(echo "[req]"; echo distinguished_name=req; echo "[san]"; echo subjectAltName=DNS:localhost,IP.1:192.168.0.10) -subj /CN=*
echo "done"
