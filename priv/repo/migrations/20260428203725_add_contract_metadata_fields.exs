defmodule Frontend.Repo.Migrations.AddContractMetadataFields do
  use Ecto.Migration

  def change do
    alter table(:contracts) do
      add :measurement_unit, :string
      add :trading_period_start, :date
      add :trading_period_end, :date
      add :measurement_period_start, :date
      add :measurement_period_end, :date
      add :data_provider_name, :string
      add :data_provider_station_mode, :string
      add :description, :text
    end
  end
end
