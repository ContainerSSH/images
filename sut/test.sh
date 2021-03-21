#!/bin/sh

set -euo pipefail

RESULTS_FILE=/tmp/$$.results
EXPECTED_FILE=/tmp/$$.expected

trap '{ rm -rf -- "${RESULTS_FILE}" "${EXPECTED_FILE}"; }' EXIT

echo -n "Waiting for ContainerSSH to become available"
OPEN=0
TRIES=0
while [ ${OPEN} -eq 0 ]; do
  if [ "${TRIES}" -gt 30 ]; then
    echo "failed."
    exit 1
  fi
  sleep 1
  echo -n "."
  OPEN=$(nmap containerssh -PN -p 2222 2>/dev/null | egrep 'open' | wc -l)
  TRIES=$(($TRIES + 1))
done
echo "done."

echo -n 'Hello world!' >${EXPECTED_FILE}
set +e
sshpass -p 'bar' ssh foo@containerssh -p 2222 -o StrictHostKeyChecking=no "echo -n 'Hello world!'" >${RESULTS_FILE}
EXIT_CODE=$?
set -e
if [ "${EXIT_CODE}" -ne 0 ]; then
  echo "Unexpected exit code from ContainerSSH: ${EXIT_CODE}" >&2
  exit ${EXIT_CODE}
fi
if [ "$(diff -NaurZw ${EXPECTED_FILE} ${RESULTS_FILE} | wc -l)" -ne 0 ]; then
  echo "SSH diff test failed." >&2
  diff -NaurZw ${EXPECTED_FILE} ${RESULTS_FILE}
  exit 1
fi

echo "Test successful."