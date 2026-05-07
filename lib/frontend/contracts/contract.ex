defmodule Frontend.Contracts.Contract do
  use Ecto.Schema

  schema "contracts" do
    field(:name, :string)
    field(:region, :string)
    field(:metric, :string)
    field(:status, :string)
    field(:threshold, :integer)
    field(:multiplier, :integer)

    field(:measurement_unit, :string)
    field(:trading_period_start, :date)
    field(:trading_period_end, :date)
    field(:measurement_period_start, :date)
    field(:measurement_period_end, :date)
    field(:data_provider_name, :string)
    field(:data_provider_station_mode, :string)
    field(:description, :string)

    timestamps()
  end
end
