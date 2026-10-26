# Workbench Display — Setup & Running Guide

**For the person setting this up at the plant.** No programming needed. Just
follow the steps in order. If anything doesn't work, jump to the
**Troubleshooting** section near the end, or call the contact listed at the
bottom.

---

## 1. What this system does (in one minute)

When an operator scans the barcode label on a part with the **Zebra DS3678**
scanner, the correct assembly specification **picture** appears full-screen on a
**second monitor** at the workbench.

The picture **stays on the screen until the next part is scanned** — it never
times out or goes blank. Scan the next part and the picture switches to that
part.

The **same scanner keeps working with your existing software too.** The scanner
plugs into a **serial (COM) port**, and a small **serial splitter** copies every
scan to two places at once: your current program **and** this display. Your
existing setup keeps working exactly as before.

Everything runs on **one PC**. There is no internet needed and no cloud — it all
works offline, right on that computer.

---

## 2. What you need before starting

Tick these off first:

- [ ] The **PC** (Windows) that will run everything.
- [ ] **Two monitors** connected to that PC: the **main** screen (where your
      existing software runs) and a **second** screen (where the part pictures
      will show).
- [ ] The **Zebra DS3678** scanner, in **serial mode**, connected to the PC's
      **serial (COM) port** (through its cradle / RS-232 or USB-to-serial lead).
- [ ] The **serial splitter** already installed and configured (this was set up
      for you — see the note in Step 3.4). You just need to know **which COM
      port number** is assigned to *this display*.
- [ ] Your **existing scanning software**, if you have one.
- [ ] **Google Chrome** installed on the PC. (Free from google.com/chrome.)
- [ ] The **`workbench-display` folder** (the one this guide is inside).

---

## 3. One-time setup

You only do Section 3 **once**. After that, starting it each day is just one
double-click (Section 4).

### Step 3.1 — Copy the folder onto the PC

Copy the whole **`workbench-display`** folder to somewhere simple, like:

```
C:\workbench-display
```

Keep the folder together — don't move individual files out of it.

### Step 3.2 — Set up the two monitors

1. Right-click on the desktop → **Display settings**.
2. You'll see two boxes labelled **1** and **2**. Drag them so they match how
   the monitors physically sit on the desk (left/right).
3. Click monitor box **1**, scroll down, and note whether **"Make this my main
   display"** is ticked. That one is the **main** monitor (your existing
   software). The **other** one (usually **2**) is where the part pictures will
   appear.

### Step 3.3 — Tell the system which monitor to use

The system needs to know where the second monitor is. Do this:

1. Click the **Start** menu, type **PowerShell**, and open it.
2. Copy-paste this line and press Enter:

   ```
   Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.Screen]::AllScreens | ForEach-Object { '{0}  Primary={1}  X={2}  Y={3}  Size={4}x{5}' -f $_.DeviceName,$_.Primary,$_.Bounds.X,$_.Bounds.Y,$_.Bounds.Width,$_.Bounds.Height }
   ```

3. You'll get two lines. Find the line that says **`Primary=False`** — that's
   the second monitor. Note its **X** and **Y** numbers. Example:

   ```
   \\.\DISPLAY2  Primary=False  X=1920  Y=0  Size=1920x1080
   ```
   Here X = **1920** and Y = **0**.

4. Now open the file **`start.bat`** in the folder: right-click it → **Edit**
   (it opens in Notepad). Near the top you'll see:

   ```
   set MONITOR_X=1920
   set MONITOR_Y=0
   ```

   Change those two numbers to the **X** and **Y** you just noted. Save and
   close.

   > Tip: the X number can be negative (like `-1920`) if the second monitor is
   > on the **left**. That's fine — type it exactly as shown, minus sign and all.

### Step 3.4 — The scanner, the COM port, and the splitter

The DS3678 sends its scans over a **serial (COM) port** — not as a keyboard.
Because a COM port can only be opened by **one** program at a time, a **serial
splitter** was installed so the scan reaches **both** your existing software
**and** this display:

```
  DS3678  ─►  physical COM port  ─►  splitter  ─┬─►  your existing software
                                                └─►  this display (serial-capture)
```

The splitter (com0com + hub4com) creates extra **virtual COM ports**. One of
them is dedicated to **this display**. You need that port number.

**Tell the display which port and speed to use:**

1. Open **`config.json`** in the folder (right-click → **Edit**).
2. Find these three lines and set them:

   ```
   "serial_port": "COM3",
   "serial_mode": "baud=9600 parity=N data=8 stop=1",
   "serial_debug": false,
   ```

   - **`serial_port`** — the **virtual COM port assigned to this display** by the
     splitter (ask whoever set up the splitter; commonly `COM3`/`COM4`/`COM10`…).
   - **`serial_mode`** — must match how the DS3678 is configured. `baud=9600`
     is the DS3678's default; if your scanner was set to a different speed (e.g.
     `115200`), change the number to match.
   - Leave **`serial_debug`** as `false` for now (Step 3.5 uses it if needed).

3. Save and close.

> **Where does that virtual COM port come from?** From a one-time "serial
> splitter" setup. If it hasn't been done yet, whoever configures it should
> follow **Appendix A** at the end of this guide — it explains the whole thing
> in plain language, step by step. Once it's set up, they'll tell you the port
> number to put in `serial_port` here.

### Step 3.5 — Test that the display sees the scanner

1. Double-click **`start.bat`** (see Section 4 for what to expect).
2. When the **"Scan Part to Begin"** screen is up on the second monitor, **scan a
   part label** with the DS3678.
3. **What you should see:** the correct part **picture** appears on the second
   monitor.

If the picture **appears** → 👍 you're done, skip to Section 4.

If **nothing happens**, turn on debug to see what's arriving:

1. Edit **`config.json`**, set `"serial_debug": true`, save.
2. Double-click `start.bat` again. A small black window titled
   **serial-capture** stays open.
3. Scan a label and watch that window:
   - **You see `raw ... bytes:` and `received: "...P82301..."`** → the port is
     right; the display should be updating. If it isn't, the part number may not
     be in the list (see Troubleshooting → *Unknown Part*).
   - **You see `cannot open COMx`** → the **`serial_port` number is wrong**, or
     another program is holding the port (the splitter isn't routing to us).
     Fix the number in `config.json` / check the splitter.
   - **The window stays blank when you scan** → the scanner data isn't reaching
     this port at all: check the DS3678 is in **serial output** mode and the
     splitter is running and pointed at this port. A **baud mismatch** can also
     show as garbled bytes — set `serial_mode` to match the scanner.
4. When it's working, set `"serial_debug": false` again and restart.

### Step 3.6 — (Optional) Make it start automatically when the PC turns on

If you want the system to launch by itself every time the PC boots:

1. Right-click **`install.bat`** → **Run as administrator**.
2. Say **Yes** to the security prompt.

(To undo this later, run **`uninstall.bat`** the same way.)

---

## 4. Starting it each day

**Double-click `start.bat`** in the folder.

(If you set up auto-start above, it starts on its own when the PC turns on.)

**What you should see, in order:**

1. Two small black windows appear and **shrink to the taskbar** — the **engine**
   and the **serial-capture** reader. Leave them running; don't close them.
2. A few seconds later, the **second monitor** goes full-screen and shows a
   **"Scan Part to Begin"** waiting screen.

The system is now ready. Operators just use your existing software as normal —
the picture on the second monitor follows along by itself. **No window needs to
be in front** — scans arrive over the serial port, not the keyboard.

---

## 5. How the operator uses it

1. **Scan** the part label the way you normally do → the correct spec **picture**
   appears on the second monitor, and your existing software receives the scan
   as usual.
2. Assemble the part using the picture. **The picture stays on screen the whole
   time** — it does not time out or disappear.
3. When the next part comes, **scan its label** → the picture switches to that
   part.

Good to know:

- Scanning the **same** label again just re-shows that same picture (it briefly
  refreshes) — nothing is skipped.
- Scanning a **different** part's label switches to that part's picture.
- Scanning a label the system doesn't recognise shows a red **error** screen
  with a beep for a moment, then it goes back to the picture that was showing.

---

## 6. Shutting down / restarting

- **To shut down:** press **Alt + F4** on the part-picture screen to close it,
  then close the small black window(s) in the taskbar. (Or just shut down the PC
  normally — that closes everything.)
- **To restart if something looks stuck:** close everything as above, then
  double-click **`start.bat`** again.

---

## 7. Troubleshooting

| Problem | What to check / do |
|---|---|
| **My existing software gets the scan, but the second-monitor picture doesn't change** | The **serial-capture** reader may not be running, or it's on the wrong port. Look for its black window in the taskbar; if it's missing, double-click `start.bat`. Then do Step 3.5 with `serial_debug` on to see whether scans are reaching our port. |
| **`serial-capture` window says `cannot open COMx`** | Either the `serial_port` number in `config.json` is wrong, or another program already holds that port (the splitter must give *us* a dedicated virtual port). Fix the number / check the splitter, then restart. |
| **Debug window stays blank when I scan** | The scan isn't reaching this port. Check: the DS3678 is in **serial output** mode; the splitter is running and routing the scanner to our port; the **baud** in `serial_mode` matches the scanner. |
| **Debug window shows garbled/odd bytes** | Usually a **baud mismatch**. Set `serial_mode` (e.g. `baud=115200 …`) to match how the DS3678 is configured, and restart. |
| **"Unknown Part" red screen when I scan a part** | The 5-digit part number isn't in the list yet, or it's a different label than expected. Note the number and contact the person below. |
| **The picture shows on the wrong monitor** | The monitor numbers in `start.bat` need adjusting. Redo Step 3.3 with the correct X/Y for the second monitor. |
| **The waiting screen never appears after start.bat** | Make sure **Google Chrome** is installed, and that the black **engine** window is still running (minimised). If it closed, double-click `start.bat` again. |
| **Everything is frozen** | Close everything, double-click `start.bat`. If still stuck, restart the PC. |

---

## 8. For the plant's IT person — pictures, parts & serial settings

- **Pictures** are ordinary **PNG files** in the **`assets`** folder. To replace
  one, drop a new PNG in with the **same file name** and restart with `start.bat`.
- **Which part shows which picture**, and the **serial port / speed**, live in a
  text file called **`config.json`**:
  - `parts` maps each **5-digit part code** (e.g. `82301`) to its picture and
    description. The system reads the 5 digits after the `P` in the scanned label
    (a full label looks like `…P82301GB000MQC…`) to pick the picture.
  - `serial_port`, `serial_mode`, `serial_debug` — the COM port settings from
    Step 3.4.
- To **add or change** a part mapping, please contact the person below rather
  than guessing, so the file stays valid.

Nothing here needs the internet, and no data leaves the PC. A record of every
scan the display received is saved automatically in the **`logs`** folder (one
file per day).

---

## Appendix A — Setting up the serial splitter (one-time)

**Who this is for:** the person configuring the PC. No programming — you'll
install a small free tool and run one command. Takes about 15 minutes.

### The idea, in plain words

The scanner sends its data through **one** cable/port (a "COM port"). The
problem: **only one program can listen to a COM port at a time.** But we need
**two** programs to hear every scan:

- the plant's **existing software**, and
- our **spec-picture display**.

So we add a **splitter** — think of it like a **TV signal splitter** that takes
one cable and copies the picture to two TVs. Here, it copies the scanner's data
into two "pretend cables" so both programs get every scan at the same time.

We use one free tool for this — **com0com** (it comes with a helper called
**hub4com**):

- **com0com** makes **"pretend cables"** inside Windows — pairs of fake COM
  ports joined together. Whatever goes in one end comes out the other end.
- **hub4com** is the **copier** — it takes the real scanner's data and pushes a
  copy into each pretend cable.

Picture of what we're building (your COM numbers will differ):

```
                              ┌────────────────► COM10 ──► THEIR software
  DS3678 ─► COM3 ─► hub4com ──┤   (the copier)
   (real)  (real)             └────────────────► COM12 ──► our display (serial-capture)
```

### Step-by-step

**1. Install the tool.**
Download **com0com** (the "signed" version, which includes **hub4com**) and run
its installer. Accept the driver-install prompt from Windows. It installs to
something like `C:\Program Files (x86)\com0com\`.

**2. Find the scanner's real COM port.**
Plug in the DS3678 (serial mode). Open **Device Manager** (Start → type "Device
Manager") → expand **Ports (COM & LPT)**. Find the scanner / USB-serial adapter
and note its number, e.g. **COM3**. (Unplug/replug to see which one appears.)

**3. Make two "pretend cables" with com0com.**
Open **com0com → Setup** (Start menu → "Setup for com0com"). Click **Add Pair**
twice so you have two pairs. For each of the **four** ends, tick
**"use Ports class"** and give it an easy number in the **COM#** box, so you end
up with:

| Pretend cable | Front end (a program listens here) | Back end (hub4com feeds here) |
|---|---|---|
| Cable 1 | **COM10** | COM11 |
| Cable 2 | **COM12** | COM13 |

Click **Apply**. (If those numbers are taken, pick any free ones and just keep
track of which is which.)

**4. Run the copier (hub4com).**
Open **Command Prompt as administrator** and run (all on one line):

```
"C:\Program Files (x86)\com0com\hub4com.exe" --octs=off --route=0:1,2 \\.\COM3 \\.\COM11 \\.\COM13
```

In plain words: *"take everything coming from **COM3** (the scanner) and send a
copy into **COM11** and **COM13**."* Those copies come out at **COM10** and
**COM12**, where the two programs listen.

- Replace **COM3** with your scanner's real port from Step 2.
- Replace **COM11 / COM13** with your two **back-end** ports from Step 3.
- Leave the window open — hub4com must **keep running** for scans to flow. (To
  make it permanent, see "Make it automatic" below.)

**5. Point each program at its own front-end port.**

- **Their software:** set it to read **COM10** (Cable 1's front end).
- **Our display:** open `config.json` and set **`"serial_port": "COM12"`**
  (Cable 2's front end).
- **Match the speed everywhere:** the DS3678's baud, `serial_mode` in
  `config.json`, and (if asked) hub4com must all agree — default **9600**.

**6. Test it.**
- In `config.json` set `"serial_debug": true`, then run **`start.bat`**.
- Scan a label. In the **serial-capture** black window you should see
  `received: "...P82301..."`, and the spec picture should appear. Their software
  should show the same scan on the main monitor.
- When happy, set `"serial_debug": false` and restart.

### Make it automatic (so it survives a reboot)

hub4com must run every time the PC starts. Easiest way:

1. Open **Task Scheduler** → **Create Basic Task** → name it "Scanner splitter".
2. Trigger: **When the computer starts**.
3. Action: **Start a program** → Program: the full `hub4com.exe` path above,
   and put the `--octs=off --route=0:1,2 \\.\COM3 \\.\COM11 \\.\COM13` part in
   **"Add arguments"**.
4. Tick **"Run whether user is logged on or not"** and **"Run with highest
   privileges"**. Finish.

### Quick checklist

- [ ] com0com installed; **two** pairs created (COM10/COM11 and COM12/COM13).
- [ ] Scanner's real port known (e.g. COM3).
- [ ] hub4com running: real port → both **back-end** ports (COM11, COM13).
- [ ] Their software reads **COM10**; our `serial_port` = **COM12**.
- [ ] Same baud everywhere.
- [ ] Both programs show the scan when you test.

> If a scan reaches only one program: the app that's missing it is either reading
> the wrong port, or still holding the **real** port directly (COM3) instead of
> its pretend-cable front end — point it at its COMxx front end and retry.

---

## Appendix B — If the splitter fights you (READ THIS FIRST)

Setting up com0com + hub4com is fiddly. Here's the important thing:

> ### 🟢 The DEMO does NOT need the splitter at all.
> The splitter only exists so the plant's **own** software can read the scanner
> **at the same time** as this display. For a demo, only this display needs the
> scans — so just point it **straight at the scanner** and skip com0com/hub4com
> entirely. **Nothing to install, nothing to get blocked.**

### Path 1 — Demo (simplest, recommended for the first working run)

1. Find the scanner's real COM port in **Device Manager → Ports (COM & LPT)**
   (in the last setup it was **COM20**).
2. Close the plant's own scanner software (so the port is free).
3. In `config.json` set:
   ```
   "serial_port": "COM20",
   "serial_mirror_port": "",
   ```
4. Run `start.bat`, scan a part → the sheet appears. **Done.** No com0com, no hub4com.

> **Baud note:** `mode COM20` reported the scanner as **1200 baud, 7 data bits,
> No parity, 1 stop** — so config now uses `baud=1200 parity=N data=7 stop=1`.
> If a scan shows up **garbled** in the `serial-capture` window (turn on
> `"serial_debug": true`), the scanner is actually at a different speed — change
> `serial_mode` to `baud=9600 parity=N data=8 stop=1` and try again.

### Path 2 — Production sharing, WITHOUT hub4com (recommended)

This display can now **copy the scans to the plant's software itself**, so you
**don't need hub4com** (the program that got "Access denied"). You only need
**one** com0com pair.

1. In **com0com Setup**, make one *connected* pair (this is com0com's normal
   default), e.g. **COM1 ↔ COM11**. (You do **not** need two separate pairs, and
   you do **not** need hub4com.)
2. In `config.json`:
   ```
   "serial_port": "COM20",          <- the real scanner
   "serial_mirror_port": "COM11",   <- our reader writes a copy here
   ```
3. Point the **plant's own software at COM1** (the other end of the pair). Set it
   to the same baud as the scanner.
4. Run `start.bat`. Our reader reads COM20, shows the sheet, **and** streams a
   copy out COM11 → COM1 for the plant's software. One scan, both programs.

### Path 3 — If you still want to use hub4com: fixing "Access denied"

In the debug log, `hub4com.exe` failed with **"Access is denied"** on *every*
command — even `dir hub4com.exe` and `hub4com.exe --version`. That means the
problem was **not the COM ports** (those were set up correctly). **Windows
blocked the `hub4com.exe` file itself.** To fix:

1. In File Explorer, right-click **`C:\hub4com\hub4com.exe` → Properties**. On the
   **General** tab, if you see *"This file came from another computer…"* with an
   **Unblock** checkbox, tick it → **Apply**.
2. If antivirus / Windows Defender quarantined it, **restore it** and add an
   **exclusion** for `C:\hub4com\`.
3. Re-extract hub4com from its ZIP to a simple folder like `C:\hub4com\`, and run
   the command from an **Administrator** Command Prompt.
4. The command that worked once was:
   ```
   hub4com.exe --route=0:1,2 --octs=off --baud=9600 \\.\COM20 \\.\CNCB2 \\.\CNCB3
   ```
   (This needs the **two-pair** com0com layout from the log. Path 2 above avoids
   all of this.)

**Bottom line:** for the demo use **Path 1**; for the real line use **Path 2**.
Only fall back to hub4com (Path 3) if you specifically want it.

---

## 9. Who to contact

If you get stuck on anything in this guide:

- **Name:** _______________________
- **Phone / email:** _______________________

Please include a photo of the screen and the part label if something looks
wrong — it makes it much faster to help.
