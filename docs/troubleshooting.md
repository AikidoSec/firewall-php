# Troubleshooting

## Check logs for errors

`cat /var/log/aikido-*/*`

## Check if Aikido module has enabled

`php -i | grep "aikido support"`

Expected output: `aikido support => enabled`
