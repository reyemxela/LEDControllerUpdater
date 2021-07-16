#!/bin/env python3

import requests
import zipfile
import tempfile
import os
import subprocess
import tkinter as tk
from tkinter import ttk

FIRMWARE_REPO = 'https://github.com/wingnut-tech/FT-Night-Radian-LED-Controller'
UPDATER_REPO  = 'https://github.com/reyemxela/LEDControllerUpdater'
API_URL       = 'https://api.github.com/repos/wingnut-tech/FT-Night-Radian-LED-Controller/releases'

TMP_PATH = tempfile.gettempdir()

class MainApplication(tk.Frame):
  def __init__(self, master=None):
    super().__init__(master)
    self.pack()
    self.create_widgets()

    self.releases = None
    self.update_releases()


  def create_widgets(self):
    # row 0
    self.labelTop = tk.Label(self, text='WingnutTech LED Controller Updater')
    self.labelTop.grid(padx=30, pady=10, row=0, column=0, columnspan=2)
    
    # row 1
    self.ch340Button = tk.Button(self, text='Install CH340 Drivers', command=self.install_ch340)
    self.ch340Button.grid(pady=20, row=1, padx=10, column=0, columnspan=2) #, sticky='w')

    # row 2
    self.versionLabel = tk.Label(self, text='Version: ')
    self.versionLabel.grid(padx=10, pady=10, row=2, column=0, sticky='e')

    self.versionCombo = ttk.Combobox(self, state='readonly', width=10)
    self.versionCombo.bind('<<ComboboxSelected>>', self.update_layouts)
    self.versionCombo.grid(pady=10, row=2, column=1, sticky='w')

    # row 3
    self.layoutLabel = tk.Label(self, text='Layout: ')
    self.layoutLabel.grid(padx=10, pady=10, row=3, column=0, sticky='e')

    self.layoutCombo = ttk.Combobox(self, state='readonly', width=20)
    self.layoutCombo.grid(pady=10, row=3, column=1, sticky='w')

    # row 4
    self.flashButton = tk.Button(self, text='Download and Flash', command=self.flash_firmware)
    self.flashButton.grid(pady=30, row=4, column=0, columnspan=2)


  def update_releases(self):
    r = requests.get(API_URL)
    if r:
      self.releases = {}
      for version in r.json():
        assets = {}
        for asset in version['assets']:
          if asset['name'].endswith('.hex'):
            assets[asset['name']] = asset['browser_download_url']
        self.releases[version['name']] = assets or {'-- No files found --': None}

      self.versionCombo['values'] = list(self.releases.keys())
      self.versionCombo.current(0)
      self.update_layouts()


  def update_layouts(self, *event):
    self.layoutCombo['values'] = list(self.releases[self.versionCombo.get()].keys())
    self.layoutCombo.current(0)


  def install_avrdude(self):
    if not os.path.exists(f'{TMP_PATH}/avrdude/avrdude.exe'):
      filename = self.download_file(UPDATER_REPO, 'v1.0.0', 'avrdude.zip')
      if filename:
        dirname = self.unzip_file(filename, 'avrdude')


  def install_ch340(self):
    if True:
      filename = self.download_file(UPDATER_REPO, 'v1.0.0', 'CH34x_Install_Windows_v3_4.zip')
      if filename:
        dirname = self.unzip_file(filename)
        if dirname:
          subprocess.run(f'{dirname}/CH34x_Install_Windows_v3_4.EXE')


  def download_file(self, repo, version, filename):
    try:
      path = f'{TMP_PATH}/{filename}'
      r = requests.get(f'{repo}/releases/download/{version}/{filename}')
      open(f'{path}', 'wb').write(r.content)

      return path
    except:
      return None


  def unzip_file(self, filename, dirname=None):
    try:
      dirname = dirname or os.path.basename(filename).removesuffix('.zip')
      dirname = f'{TMP_PATH}/{dirname}'

      z = zipfile.ZipFile(filename)
      z.extractall(dirname)

      os.remove(filename)

      return dirname
    except:
      return None


  def flash_firmware(self):
    ver = self.versionCombo.get()
    hexfile = self.layoutCombo.get()
    if ver and hexfile:
      if not os.path.exists(f'{TMP_PATH}/{hexfile}'):
        self.download_file(FIRMWARE_REPO, ver, hexfile)
      try:
        self.install_avrdude()
        avrdude_command = f'"{avrdudepath}/avrdude.exe" -C{avrdudepath}/avrdude.conf -v -carduino -patmega328p -b115200 -D -Uflash:w:{hexfile}:i'
        return True
      except:
        return False


if __name__ == '__main__':
  app = MainApplication()
  app.mainloop()
