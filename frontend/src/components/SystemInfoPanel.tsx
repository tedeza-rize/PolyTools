import { useEffect, useState } from "react";
import { Card } from "@fluentui/react-components";
import { PolyToolsService } from "../../bindings/polytools/internal/services";
import { ModuleInfo } from "../types";
import { useT } from "../i18n";

type Stats = {
  os: string;
  cpuName: string;
  cpuPercent: number;
  memTotalBytes: number;
  memUsedBytes: number;
  memLoadPercent: number;
  gpuName: string;
  uptimeSeconds: number;
  hasBattery: boolean;
  batteryPercent: number;
  batteryOnAC: boolean;
};

function gb(b: number) {
  return (b / 1073741824).toFixed(1);
}

function fmtUptime(sec: number) {
  const d = Math.floor(sec / 86400);
  const h = Math.floor((sec % 86400) / 3600);
  const m = Math.floor((sec % 3600) / 60);
  if (d > 0) return `${d}d ${h}h ${m}m`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

function Stat({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <Card size="small" className="stat-card">
      <div className="stat-label">{label}</div>
      <div className="stat-value">{value}</div>
      {sub && <div className="stat-sub" title={sub}>{sub}</div>}
    </Card>
  );
}

export function SystemInfoPanel({ module }: { module: ModuleInfo }) {
  const t = useT();
  const [stats, setStats] = useState<Stats | null>(null);

  const refreshMs =
    Math.max(
      1,
      Number(module.settings?.find((f) => f.key === "refreshSeconds")?.value ?? 2)
    ) * 1000;

  useEffect(() => {
    if (!module.enabled) {
      setStats(null);
      return;
    }
    let stop = false;
    const poll = () =>
      PolyToolsService.SystemInfo()
        .then((s) => {
          if (!stop) setStats(s);
        })
        .catch(console.error);
    poll();
    const id = setInterval(poll, refreshMs);
    return () => {
      stop = true;
      clearInterval(id);
    };
  }, [module.enabled, refreshMs]);

  if (!module.enabled || !stats) return null;

  return (
    <div className="stat-grid">
      <Stat
        label={t("stat.cpu")}
        value={`${stats.cpuPercent.toFixed(1)}%`}
        sub={stats.cpuName}
      />
      <Stat
        label={t("stat.memory")}
        value={`${stats.memLoadPercent}%`}
        sub={`${gb(stats.memUsedBytes)} / ${gb(stats.memTotalBytes)} GB`}
      />
      {stats.gpuName && <Stat label={t("stat.gpu")} value="" sub={stats.gpuName} />}
      <Stat label={t("stat.os")} value="" sub={stats.os} />
      <Stat label={t("stat.uptime")} value={fmtUptime(stats.uptimeSeconds)} />
      {stats.hasBattery && (
        <Stat
          label={t("stat.battery")}
          value={`${stats.batteryPercent}%`}
          sub={stats.batteryOnAC ? t("stat.ac") : t("stat.onBattery")}
        />
      )}
    </div>
  );
}
