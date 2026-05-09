import { useEffect, useRef } from "react";
import { destroyChart, renderChart } from "../lib/chart";

export function MarketChart({ contract, chartData, now, onConfigChange }) {
  const shellRef = useRef(null);
  const canvasRef = useRef(null);

  useEffect(() => {
    const shell = shellRef.current;
    if (!shell) return undefined;

    const handleConfigChange = (event) => {
      const detail = event.detail || {};
      onConfigChange?.({
        lookbackSeconds: Number(detail.lookbackSeconds),
        bucketSeconds: Number(detail.bucketSeconds),
      });
    };

    shell.addEventListener("chart-config-change", handleConfigChange);
    return () => shell.removeEventListener("chart-config-change", handleConfigChange);
  }, [onConfigChange]);

  useEffect(() => {
    if (!canvasRef.current || !contract || !chartData?.points?.length) return;
    renderChart(canvasRef.current, contract, chartData, now);
  }, [contract, chartData, now]);

  useEffect(
    () => () => {
      if (canvasRef.current) {
        destroyChart(canvasRef.current);
      }
    },
    [],
  );

  return (
    <div className="panel--chart">
      <div className="chart-shell" ref={shellRef}>
        <canvas id="chart-canvas" ref={canvasRef} />
      </div>
    </div>
  );
}
