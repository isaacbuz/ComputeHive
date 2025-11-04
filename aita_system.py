# aita_system.py
# ---------------------------------------------------------
# Automated Intelligent Technical Analysis (AITA) — MVP
# One-file reference implementation:
# - Data ingest (yfinance)
# - Indicators & features (ATR/RSI/MACD, volume z-score)
# - Pivot detection (zigzag-lite)
# - Pattern rules: Cup-and-Handle, Bull Flag
# - Fib retracements (pivot-anchored) + optional extensions
# - Anchored VWAP (event/swing anchored)
# - VPVR (volume-by-price bins) → HVN/LVN shelves
# - S/R levels (pivot clustering + touches)
# - Confluence scoring
# - Signal compiler → plan JSON (ENTRY/SL/TP/targets/quality)
# - Options planner (IV rank → structure + DTE + strikes)
# - Overlay spec (primitives for a frontend renderer)
# - FastAPI endpoint /analyze
#
# Dependencies:
#   pip install yfinance pandas numpy scipy ta fastapi uvicorn pydantic
#
# NOTE: This is an educational MVP—not financial advice.

from __future__ import annotations
import math, statistics, json, datetime as dt
from dataclasses import dataclass, asdict
from typing import List, Dict, Optional, Tuple

import numpy as np
import pandas as pd
import yfinance as yf
from scipy.signal import argrelextrema
from scipy.stats import zscore
from ta.volatility import AverageTrueRange, BollingerBands
from ta.momentum import RSIIndicator
from ta.trend import MACD

# ---------- Utilities

def atr(df: pd.DataFrame, n=14) -> pd.Series:
    return AverageTrueRange(df["High"], df["Low"], df["Close"], n).average_true_range()


def macd(df: pd.DataFrame) -> Tuple[pd.Series, pd.Series, pd.Series]:
    m = MACD(df["Close"], window_slow=26, window_fast=12, window_sign=9)
    return m.macd(), m.macd_signal(), m.macd_diff()


def rsi(df: pd.DataFrame, n=14) -> pd.Series:
    return RSIIndicator(df["Close"], n).rsi()


def vol_zscore(df: pd.DataFrame, n=20) -> pd.Series:
    v = df["Volume"].astype(float)
    m = v.rolling(n).mean()
    s = v.rolling(n).std(ddof=0)
    return (v - m) / s


def typical_price(df: pd.DataFrame) -> pd.Series:
    return (df["High"] + df["Low"] + df["Close"]) / 3.0


# ---------- Anchored VWAP

def anchored_vwap(df: pd.DataFrame, anchor_idx: int) -> pd.Series:
    """
    Simple AVWAP from anchor_idx forward.
    """
    tp = typical_price(df)
    vol = df["Volume"].astype(float)
    ctp = (tp.iloc[anchor_idx:] * vol.iloc[anchor_idx:]).cumsum()
    cv = vol.iloc[anchor_idx:].cumsum()
    avwap = (ctp / cv).reindex(df.index)
    # forward fill after anchor; NaN before anchor
    return avwap


# ---------- VPVR: Volume by Price (visible range)

def vpvr(df: pd.DataFrame, bins=60) -> Dict[str, np.ndarray]:
    low = df["Low"].min()
    high = df["High"].max()
    hist_prices = np.linspace(low, high, bins + 1)
    tp = typical_price(df).values
    vol = df["Volume"].values.astype(float)
    # assign each bar's volume to its typical-price bin
    inds = np.searchsorted(hist_prices, tp, side="right") - 1
    inds = np.clip(inds, 0, bins - 1)
    vbp = np.zeros(bins)
    np.add.at(vbp, inds, vol)
    centers = (hist_prices[:-1] + hist_prices[1:]) / 2
    return {"price_centers": centers, "volume": vbp}


def hvn_lvn(vp: Dict[str, np.ndarray], prominence=1.0) -> Tuple[List[float], List[float]]:
    v = vp["volume"]
    p = vp["price_centers"]
    # local maxima/minima on the volume profile
    highs = argrelextrema(v, np.greater_equal, order=2)[0]
    lows = argrelextrema(v, np.less_equal, order=2)[0]
    # Filter by prominence-ish (relative to mean)
    meanv = v.mean() if v.size else 0
    hvns = [float(p[i]) for i in highs if v[i] >= meanv * (1 + prominence * 0.1)]
    lvns = [float(p[i]) for i in lows if v[i] <= meanv * (1 - prominence * 0.1)]
    return hvns, lvns


# ---------- ZigZag-lite pivots for swings & S/R

def find_pivots(df: pd.DataFrame, order=5) -> Tuple[List[int], List[int]]:
    """
    Returns indices of swing highs and lows using local extremas.
    order=5 → pivot must be the highest/lowest over ±5 bars.
    """
    highs = argrelextrema(df["High"].values, np.greater, order=order)[0].tolist()
    lows = argrelextrema(df["Low"].values, np.less, order=order)[0].tolist()
    return highs, lows


def cluster_levels(prices: List[float], tolerance: float = 0.003) -> List[Tuple[float, int]]:
    """
    Cluster nearby price levels (within % tolerance). Returns list of (level, touches).
    """
    if not prices:
        return []
    prices = sorted(prices)
    clusters = [[prices[0]]]
    for x in prices[1:]:
        if abs(x - clusters[-1][-1]) / clusters[-1][-1] <= tolerance:
            clusters[-1].append(x)
        else:
            clusters.append([x])
    result = [(float(sum(c) / len(c)), len(c)) for c in clusters]
    # sort by touches desc
    result.sort(key=lambda t: (-t[1], t[0]))
    return result


def sr_levels_from_pivots(df: pd.DataFrame, order=5, tolerance=0.003) -> List[Tuple[float, int]]:
    highs, lows = find_pivots(df, order)
    levels = [float(df["High"].iloc[i]) for i in highs] + [float(df["Low"].iloc[i]) for i in lows]
    return cluster_levels(levels, tolerance)


# ---------- Fib retracements (pivot-anchored)

def fib_levels_from_last_swing(df: pd.DataFrame, order=5, uptrend: Optional[bool] = None) -> Dict[str, float]:
    highs, lows = find_pivots(df, order)
    if not highs or not lows:
        return {}
    last_high = max(highs)
    last_low = max(lows)
    # decide direction based on which pivot is newer unless forced
    if uptrend is None:
        uptrend = last_low < last_high
    if uptrend:
        # last confirmed swing: low -> high
        start = max([i for i in lows if i < last_high], default=lows[-1])
        end = last_high
        lo = df["Low"].iloc[start]
        hi = df["High"].iloc[end]
        diff = hi - lo
        ratios = [0, 0.236, 0.382, 0.5, 0.618, 0.786, 1.0]
        levels = {f"{int(r * 100)}%": float(hi - r * diff) for r in ratios}
        # optional extensions
        levels.update({"127.2%": float(hi + 0.272 * diff), "161.8%": float(hi + 0.618 * diff)})
        return {"direction": "up", "start_idx": int(start), "end_idx": int(end), "levels": levels}
    else:
        # down swing: high -> low
        start = max([i for i in highs if i < last_low], default=highs[-1])
        end = last_low
        hi = df["High"].iloc[start]
        lo = df["Low"].iloc[end]
        diff = hi - lo
        ratios = [0, 0.236, 0.382, 0.5, 0.618, 0.786, 1.0]
        levels = {f"{int(r * 100)}%": float(lo + r * diff) for r in ratios}
        levels.update({"127.2%": float(lo - 0.272 * diff), "161.8%": float(lo - 0.618 * diff)})
        return {"direction": "down", "start_idx": int(start), "end_idx": int(end), "levels": levels}


# ---------- Pattern rules

@dataclass
class DetectedPattern:
    name: str
    score: float
    meta: Dict


def detect_cup_handle(df: pd.DataFrame, order=5, min_depth_pct=0.08, max_depth_pct=0.40) -> Optional[DetectedPattern]:
    highs, lows = find_pivots(df, order)
    if len(highs) < 2 or len(lows) < 1:
        return None
    # use last two rims (similar highs)
    rim2 = highs[-1]
    rim1 = max([i for i in highs[:-1] if i < rim2 - 5], default=None)
    if rim1 is None:
        return None
    rim_hi1 = float(df["High"].iloc[rim1])
    rim_hi2 = float(df["High"].iloc[rim2])
    if abs(rim_hi1 - rim_hi2) / ((rim_hi1 + rim_hi2) / 2) > 0.015:
        return None
    # cup low between rims
    mid_lows = [i for i in lows if rim1 < i < rim2]
    if not mid_lows:
        return None
    cup_low_idx = int(min(mid_lows, key=lambda i: df["Low"].iloc[i]))
    cup_low = float(df["Low"].iloc[cup_low_idx])
    rim = (rim_hi1 + rim_hi2) / 2.0
    depth = (rim - cup_low) / rim
    if not (min_depth_pct <= depth <= max_depth_pct):
        return None
    # "curvature" sanity: ensure a rounded bottom via rolling mean proximity
    window = max(3, (rim2 - rim1) // 5)
    smooth = df["Close"].rolling(window).mean()
    curve_err = float(np.mean(np.abs(smooth.iloc[rim1:rim2] - df["Close"].iloc[rim1:rim2])) / rim)
    if curve_err > 0.03:
        return None
    # handle: small declining channel after rim2
    handle_slice = df.iloc[rim2 + 1 : rim2 + 1 + 15]
    if handle_slice.shape[0] < 3:
        return None
    x = np.arange(handle_slice.shape[0])
    y = handle_slice["Close"].values
    slope, intercept = np.polyfit(x, y, 1)
    # slope negative (decline) but shallow; handle depth shallow vs cup
    handle_depth = (handle_slice["High"].max() - handle_slice["Low"].min()) / rim
    if not (-0.8 <= slope <= 0) or handle_depth > min(0.5 * depth, 0.12):
        return None
    # breakout if latest close > rim
    broke = float(df["Close"].iloc[-1]) > rim
    score = 0.6 + 0.2 * (1 - curve_err / 0.03) + 0.2 * min(1, (0.5 * depth) / max(min_depth_pct, 1e-6))
    return DetectedPattern(
        name="CupAndHandle",
        score=float(min(1.0, max(0.0, score) + (0.1 if broke else 0))),
        meta={
            "rim_level": rim,
            "rim_idx1": int(rim1),
            "rim_idx2": int(rim2),
            "cup_low_idx": cup_low_idx,
            "cup_depth_pct": float(depth),
            "handle_start": int(rim2 + 1),
            "handle_end": int(rim2 + 1 + len(handle_slice) - 1),
        },
    )


def detect_bull_flag(df: pd.DataFrame, lookback=40, min_impulse_atr=1.2) -> Optional[DetectedPattern]:
    if df.shape[0] < lookback + 10:
        return None
    sub = df.iloc[-lookback:]
    atr14 = atr(sub, 14).iloc[-1]
    # Impulse leg: recent range jump vs ATR
    impulse_range = sub["High"].iloc[-10:].max() - sub["Low"].iloc[-10:].min()
    if impulse_range < min_impulse_atr * atr14:
        return None
    # Flag: last ~6 bars forming a descending tight channel w/ volume contraction
    pull = sub.iloc[-8:]
    x = np.arange(pull.shape[0])
    slope, intercept = np.polyfit(x, pull["Close"].values, 1)
    if slope >= 0:
        return None
    volz = vol_zscore(sub, 20).iloc[-8:]
    if not (volz.mean() < 0.0):  # contraction
        return None
    # score by magnitude and contraction
    score = 0.55 + 0.25 * min(1, impulse_range / (1.5 * atr14)) + 0.2 * min(1, (0 - volz.mean()) / 2)
    return DetectedPattern(
        name="BullFlag",
        score=float(min(1.0, score)),
        meta={"flag_start_idx": int(sub.index[-8:].get_loc(sub.index[-8])), "flag_len": 8},
    )


# ---------- Confluence Scoring

def confluence_score(
    df: pd.DataFrame,
    patterns: List[DetectedPattern],
    sr_levels: List[Tuple[float, int]],
    vp: Dict[str, np.ndarray],
    volz: pd.Series,
    macd_line: pd.Series,
    rsi_line: pd.Series,
) -> float:
    s = 0.0
    # pattern contribution
    if patterns:
        s += min(0.5, sum(p.score for p in patterns))
    # volume confirmation (last bar)
    if volz.iloc[-1] > 1.0:
        s += 0.15
    # momentum confirmation
    if macd_line.iloc[-1] > 0:
        s += 0.1
    if 40 <= rsi_line.iloc[-1] <= 70:
        s += 0.05
    # VPVR confluence near price (proximity to LVN for fast move)
    price = float(df["Close"].iloc[-1])
    lvns = hvn_lvn(vp)[1]
    if lvns:
        nearest_lvn = min(lvns, key=lambda x: abs(x - price))
        if abs(nearest_lvn - price) / price <= 0.01:
            s += 0.1
    # S/R touch richness
    if sr_levels and sr_levels[0][1] >= 3:
        s += 0.05
    return float(max(0.0, min(1.0, s)))


# ---------- Options planner (IV-aware, heuristic)

@dataclass
class OptionsPlan:
    structure: str
    dte: int
    strikes: Dict[str, float]
    notes: str


def iv_rank_from_hist_impvol_or_proxy(ticker: str, df: pd.DataFrame) -> float:
    """
    Try to approximate IV rank:
    - If yfinance provides option implied vol snapshots, great (limited).
    - Otherwise proxy with 20d historical volatility percentile over 1y.
    Returns 0..1
    """
    try:
        # Proxy HV percentile
        logret = np.log(df["Close"]).diff()
        hv20 = logret.rolling(20).std() * np.sqrt(252)
        hv = hv20.dropna()
        if len(hv) < 60:
            return 0.3
        cur = hv.iloc[-1]
        lo, hi = hv.min(), hv.max()
        if hi - lo < 1e-9:
            return 0.3
        pct = float((cur - lo) / (hi - lo))
        return pct
    except Exception:
        return 0.3


def choose_options_structure(iv_rank: float, bullish: bool, entry: float, tp1: float, sl: float) -> OptionsPlan:
    rr = abs(tp1 - entry) / max(1e-9, abs(entry - sl))
    if iv_rank < 0.3:
        # buy premium
        structure = "Debit Call Vertical" if bullish else "Debit Put Vertical"
        dte = 21 if rr >= 1.5 else 28
        strikes = {
            "long": round(entry * (1.02 if bullish else 0.98), 2),
            "short": round(tp1 * (1.0 if bullish else 1.0), 2),
        }
        notes = "Low IV: favor debits; place short near TP1."
    elif iv_rank > 0.6:
        structure = "Bull Put Credit Spread" if bullish else "Bear Call Credit Spread"
        dte = 21
        strikes = {
            "short": round(sl * (0.98 if bullish else 1.02), 2),
            "long": round(sl * (0.96 if bullish else 1.04), 2),
        }
        notes = "High IV: sell credit below (above) invalidation."
    else:
        structure = "Call Vertical" if bullish else "Put Vertical"
        dte = 28
        strikes = {
            "long": round(entry * (1.01 if bullish else 0.99), 2),
            "short": round(tp1, 2),
        }
        notes = "Medium IV: defined-risk vertical for balance."
    return OptionsPlan(structure, dte, strikes, notes)


# ---------- Signal compiler

def compile_trade_plan(
    df: pd.DataFrame,
    patterns: List[DetectedPattern],
    sr_levels: List[Tuple[float, int]],
    vp: Dict[str, np.ndarray],
    avwap: Optional[pd.Series],
    volz: pd.Series,
    macd_line: pd.Series,
    rsi_line: pd.Series,
) -> Dict:
    price = float(df["Close"].iloc[-1])
    rng_atr = float(atr(df, 14).iloc[-1])
    # use cup rim if available, else nearest S/R as resistance
    rim = None
    for p in patterns:
        if p.name == "CupAndHandle":
            rim = p.meta["rim_level"]
            break
    resistance = rim or (sr_levels[0][0] if sr_levels else price)
    # breakout condition
    breakout = price > (resistance + 0.25 * rng_atr)
    # entry logic: breakout or flag break; otherwise retest plan
    entry = price if breakout else resistance
    # stop: min(handle low, AVWAP - epsilon, last swing low)
    swing_lows = find_pivots(df, 5)[1]
    stop_candidates = []
    if avwap is not None and not math.isnan(avwap.iloc[-1]):
        stop_candidates.append(float(avwap.iloc[-1] - 0.1 * rng_atr))
    if swing_lows:
        stop_candidates.append(float(df["Low"].iloc[swing_lows[-1]]))
    for p in patterns:
        if p.name == "CupAndHandle":
            stop_candidates.append(float(df["Low"].iloc[p.meta["handle_start"] : p.meta["handle_end"] + 1].min()))
    if not stop_candidates:
        stop_candidates = [price - 2 * rng_atr]
    stop = float(min(stop_candidates))
    # Targets: measured move (cup depth) and Fib extension + VPVR shelf snap
    tp1, tp2 = price + 1.5 * rng_atr, price + 3 * rng_atr
    for p in patterns:
        if p.name == "CupAndHandle":
            depth = p.meta["cup_depth_pct"] * p.meta["rim_level"]
            tp1 = max(tp1, float(resistance + depth))
            tp2 = max(tp2, float(resistance + 1.5 * depth))
    hvns, lvns = hvn_lvn(vp)
    if hvns:
        # snap tp1/tp2 to nearby HVNs above price
        above = [h for h in hvns if h > price]
        if above:
            tp1 = min(above, key=lambda x: abs(x - tp1))
            farther = [h for h in above if h > tp1 * 1.02]
            if farther:
                tp2 = min(farther, key=lambda x: abs(x - tp2))
    # quality
    conf = confluence_score(df, patterns, sr_levels, vp, volz, macd_line, rsi_line)
    rr = float((tp1 - entry) / max(1e-9, entry - stop))
    dirn = "LONG" if breakout or (macd_line.iloc[-1] > 0 and rsi_line.iloc[-1] > 45) else "LONG"  # MVP targets longs
    # options plan
    ivr = iv_rank_from_hist_impvol_or_proxy("TBD", df)
    opt = choose_options_structure(ivr, bullish=(dirn == "LONG"), entry=entry, tp1=tp1, sl=stop)
    plan = {
        "symbol": "TICK",
        "timeframe": "1D",
        "thesis": ", ".join(p.name for p in patterns) or "Price Action",
        "direction": dirn,
        "entry": round(entry, 2),
        "stop": round(stop, 2),
        "targets": [round(tp1, 2), round(tp2, 2)],
        "quality": {
            "confluence": round(conf, 3),
            "volume_z": round(float(volz.iloc[-1]), 2),
            "rsi": round(float(rsi_line.iloc[-1]), 1),
            "macd": round(float(macd_line.iloc[-1]), 2),
        },
        "risk_reward": round(rr, 2),
        "options": asdict(opt),
        "notes": "Breakout confirmation requires close above resistance + volume.",
    }
    return plan


# ---------- Overlay spec (for frontend rendering)

def overlay_spec(
    df: pd.DataFrame,
    patterns: List[DetectedPattern],
    fib: Dict,
    vp: Dict[str, np.ndarray],
    plan: Dict,
    avwap: Optional[pd.Series],
) -> Dict:
    layers = []
    # pattern drawings
    for p in patterns:
        if p.name == "CupAndHandle":
            layers.append({"type": "band", "style": "rim", "y": p.meta["rim_level"], "height": 0.2 * df["Close"].iloc[-1] / 100})
            layers.append({"type": "curve", "style": "cup", "start_idx": p.meta["rim_idx1"], "mid_idx": p.meta["cup_low_idx"], "end_idx": p.meta["rim_idx2"]})
            layers.append({"type": "channel", "style": "handle", "start_idx": p.meta["handle_start"], "end_idx": p.meta["handle_end"]})
        if p.name == "BullFlag":
            layers.append({"type": "channel", "style": "flag", "start_idx": len(df) - 8, "end_idx": len(df) - 1})
    # AVWAP
    if avwap is not None:
        layers.append({"type": "line_series", "style": "avwap", "y": [None if math.isnan(y) else float(y) for y in avwap.values]})
    # VPVR shelves (right-side rendering handled in FE)
    layers.append({"type": "vpvr", "centers": vp["price_centers"].tolist(), "volume": vp["volume"].tolist()})
    # Fib levels
    if fib:
        for name, y in fib["levels"].items():
            layers.append({"type": "hline", "style": "fib", "label": name, "y": float(y)})
    # Entry/SL/TP and risk box
    layers += [
        {"type": "marker", "kind": "ENTRY", "y": plan["entry"]},
        {"type": "marker", "kind": "STOP", "y": plan["stop"]},
        {"type": "marker", "kind": "TP", "y": plan["targets"][0]},
        {"type": "marker", "kind": "TP", "y": plan["targets"][1]},
        {"type": "riskbox", "entry": plan["entry"], "stop": plan["stop"], "targets": plan["targets"]},
    ]
    return {"layers": layers}


# ---------- Orchestrator

@dataclass
class AITAConfig:
    period: str = "6mo"
    interval: str = "1d"
    pivot_order: int = 5
    vpvr_bins: int = 60


class AITA:
    def __init__(self, cfg: AITAConfig = AITAConfig()):
        self.cfg = cfg

    def analyze(self, symbol: str) -> Dict:
        data = yf.download(symbol, period=self.cfg.period, interval=self.cfg.interval, auto_adjust=True)
        if data.empty:
            raise RuntimeError(f"No data for {symbol}")
        data = data.dropna().copy()
        data["ATR"] = atr(data, 14)
        data["RSI"] = rsi(data, 14)
        m_line, m_sig, m_hist = macd(data)
        data["MACD"] = m_line
        data["VOLZ"] = vol_zscore(data, 20)

        # Patterns
        patterns: List[DetectedPattern] = []
        cnh = detect_cup_handle(data, order=self.cfg.pivot_order)
        if cnh:
            patterns.append(cnh)
        flag = detect_bull_flag(data)
        if flag:
            patterns.append(flag)

        # S/R & Fib & VPVR
        srlvls = sr_levels_from_pivots(data, order=self.cfg.pivot_order)
        fib = fib_levels_from_last_swing(data, order=self.cfg.pivot_order)
        vp = vpvr(data, bins=self.cfg.vpvr_bins)

        # AVWAP (anchor: last breakout candidate or last swing low)
        anchor_idx = None
        if cnh:
            anchor_idx = cnh.meta["rim_idx2"] + 1
        elif len(find_pivots(data, self.cfg.pivot_order)[1]) > 0:
            anchor_idx = find_pivots(data, self.cfg.pivot_order)[1][-1]
        avw = anchored_vwap(data, anchor_idx) if anchor_idx is not None else None

        # Plan
        plan = compile_trade_plan(
            data, patterns, srlvls, vp, avw, data["VOLZ"], data["MACD"], data["RSI"]
        )
        plan["symbol"] = symbol
        plan["timeframe"] = self.cfg.interval.upper()

        # Overlay
        overlay = overlay_spec(data, patterns, fib, vp, plan, avw)

        # Summary (for alerts)
        top_sr = [{"level": round(l, 2), "touches": t} for l, t in srlvls[:5]]
        summary = {
            "symbol": symbol,
            "price": round(float(data["Close"].iloc[-1]), 2),
            "patterns": [{"name": p.name, "score": round(p.score, 2)} for p in patterns],
            "top_sr_levels": top_sr,
            "plan": plan,
        }
        return {"summary": summary, "overlay": overlay}


# ---------- FastAPI

try:
    from fastapi import FastAPI
    from pydantic import BaseModel

    app = FastAPI(title="AITA API", version="0.1.0")
    analyzer = AITA()

    class AnalyzeResponse(BaseModel):
        summary: dict
        overlay: dict

    @app.get("/analyze", response_model=AnalyzeResponse)
    def analyze(symbol: str = "AVGO"):
        result = analyzer.analyze(symbol)
        return AnalyzeResponse(**result)

except Exception as e:
    # FastAPI is optional for quick script runs
    app = None


if __name__ == "__main__":
    # quick run: python aita_system.py
    a = AITA()
    out = a.analyze("AVGO")
    print(json.dumps(out["summary"], indent=2))
    # To serve API: uvicorn aita_system:app --reload
