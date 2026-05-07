alias Frontend.Repo
alias Frontend.Contracts.Contract

Repo.delete_all(Contract)

attrs = %{
  name: "Miami Jul 2026 Temperature Index",
  region: "Miami",
  metric: "Average Temperature",
  status: "Open",
  threshold: 85,
  multiplier: 10,
  measurement_unit: "deg_f",
  trading_period_start: ~D[2026-04-01],
  trading_period_end: ~D[2026-07-27],
  measurement_period_start: ~D[2026-07-01],
  measurement_period_end: ~D[2026-07-31],
  data_provider_name: "NOAA",
  data_provider_station_mode: "single_station",
  description:
    "This market references July 2026 temperature in Miami. Above tokens gain value as expected realized temperature rises above 85°F."
}

Repo.insert!(struct(Contract, attrs))
