# Serial Port & Splitter Setup (Zebra DS3678 → two programs)

**Who this is for:** the person configuring the PC. This explains how to share
**one serial scanner** between **two programs at once** — the plant's existing
software **and** the Workbench Display.

> **Before you start — do you even need a splitter?**
> A splitter is ONLY needed when the plant's own software must read the scanner
> **at the same time** as the display. **For a demo, you don't need any of this**
> — see *Method 0*.

---

## The problem in one line

A serial (COM) port can be opened by **only one program at a time**. So the
scanner on, say, **COM20** can feed either the plant software **or** the display —
not both — unless we split/copy the data.

```
                        ┌── Program A  (plant's own software)
Scanner ── COM20 ──►────┤
                        └── Program B  (Workbench Display)
```

---

## Method 0 — No splitter (use this for the DEMO)

Only the display needs the scans, so point it straight at the scanner.

1. **Device Manager → Ports (COM & LPT)** → note the scanner's port (e.g. `COM20`).
2. Close the plant's own scanner software (free the port).
3. In `config.json`:
   ```json
   "serial_port": "COM20",
   "serial_mirror_port": "",
   ```
4. Run `start.bat`, scan a part → the sheet appears. Done. No com0com, no hub4com.

**Baud:** `mode COM20` in a Command Prompt shows the scanner's current settings.
On the demo PC it showed `1200, 7, N, 1`, so config uses
`"serial_mode": "baud=1200 parity=N data=7 stop=1"`. If a scan comes through
**garbled** (turn on `"serial_debug": true` and watch the serial-capture window),
switch to `baud=9600 parity=N data=8 stop=1`.

---

## Method 1 — Built-in mirror (RECOMMENDED for production, no hub4com)

The display can copy the scans to the plant's software **itself**, so you do
**NOT need hub4com**. You only need **one** com0com virtual pair.

### Step 1 — Install com0com
Download and install **com0com** (the signed version). Accept the driver prompt.

### Step 2 — Make ONE connected virtual pair
Open **com0com → Setup** and create a normal pair, e.g. **COM1 ↔ COM11**
(com0com's default "connected pair" — what you get from a plain `install`).
Anything written to one end comes out the other.

### Step 3 — Point the display + plant software at it
In `config.json`:
```json
"serial_port": "COM20",          // the real scanner
"serial_mirror_port": "COM11",   // the display writes a copy here
```
Set the **plant's own software to read COM1** (the other end of the pair), at the
same baud as the scanner.

### Step 4 — Run
Run `start.bat`. The display reads **COM20**, shows the sheet, **and** streams a
copy out **COM11 → COM1** for the plant's software. One scan, both programs. No
hub4com anywhere.

```
Scanner COM20 ─► serial-capture ─┬─► shows the work-standard sheet
                                 └─► writes copy ─► COM11 ↔ COM1 ─► plant software
```

---

## Method 2 — Traditional com0com + hub4com

Use this only if you specifically want hub4com to do the fan-out.

### Step 1 — Two independent virtual pairs (one named end each)
In **com0com Setup**, create two ports whose *other* end is left internal:
```cmd
install PortName=COM1  PortName=-
install PortName=COM11 PortName=-
```
`list` should show (numbers may differ):
```text
CNCA2 PortName=COM1    CNCB2 PortName=-
CNCA3 PortName=COM11   CNCB3 PortName=-
```
> If the ports show in `list` but **not** in Device Manager, set them as
> Windows-visible ports: `change CNCA2 PortName=COM#` then
> `change CNCA2 RealPortName=COM1` (repeat for CNCA3/COM11), then
> **Device Manager → Action → Scan for hardware changes**.

### Step 2 — Run hub4com (it owns the physical port)
```cmd
cd /d C:\hub4com
hub4com.exe --route=0:1,2 --octs=off --baud=9600 \\.\COM20 \\.\CNCB2 \\.\CNCB3
```
This copies **COM20 → CNCB2 (→COM1) and CNCB3 (→COM11)**. Leave the window open.
Plant software reads **COM1**; set the display's `serial_port` to **COM11**
(and `serial_mirror_port` empty).

### ⚠️ If hub4com says "Access denied"
This is the failure that killed the demo. If **every** hub4com command — even
`hub4com.exe --version` or `dir hub4com.exe` — says *"Access is denied"*, the
**COM ports are fine; Windows blocked the `hub4com.exe` file**:

1. Right-click **`C:\hub4com\hub4com.exe` → Properties**. On the **General** tab,
   if there's *"This file came from another computer…"* with an **Unblock** box,
   tick it → **Apply**.
2. If **Windows Defender / antivirus** quarantined it, **restore** it and add an
   **exclusion** for `C:\hub4com\`.
3. Re-extract hub4com from its ZIP to `C:\hub4com\`, run from an **Administrator**
   Command Prompt.

*(Method 1 avoids hub4com entirely, so it can't be blocked this way.)*

---

## Verifying it works

1. Set `"serial_debug": true` in `config.json`, run `start.bat`.
2. Scan a part. In the **serial-capture** window you should see
   `received: "...P82302..."` and the sheet should appear.
3. If you also set up sharing (Method 1/2), confirm the **plant's software** shows
   the same scan.
4. Set `"serial_debug": false` again for normal use.

| Symptom | Cause / fix |
|---|---|
| `cannot open COMx` | Wrong `serial_port`, or another program holds it. Fix the number / close the other app. |
| Debug window blank on scan | Scanner data not reaching the port — check the scanner is in serial mode and (for sharing) the splitter is running. |
| Garbled bytes | Baud mismatch — set `serial_mode` to match the scanner (try 9600 8N1 vs 1200 7N1). |
| hub4com "Access denied" | Windows blocked the exe — see the box above (or switch to Method 1). |
