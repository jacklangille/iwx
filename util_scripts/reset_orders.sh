#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

mix run -e '
alias Frontend.Repo
alias Frontend.Exchange.Order
alias Frontend.Exchange.MarketSnapshot

{snapshot_count, _} = Repo.delete_all(MarketSnapshot)
{order_count, _} = Repo.delete_all(Order)

IO.puts("Deleted #{snapshot_count} market snapshots")
IO.puts("Deleted #{order_count} orders")
'