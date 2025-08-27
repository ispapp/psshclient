# ISP App Client

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Fyne](https://img.shields.io/badge/Fyne-00A8E8?style=for-the-badge&logo=fyne&logoColor=white)

A powerful and intuitive desktop application for managing network devices with ease. Built with Go and the Fyne toolkit, this app provides a seamless experience for network administrators and enthusiasts.

![App Screenshot](./Screenshot.png)  <!-- Replace with your actual screenshot -->

## ✨ Features

- **Device Discovery:** Automatically scan your network to find devices.
- **SSH Terminal:** Open an SSH terminal to any connected device.
- **Multi-Device Scripting:** Run scripts on multiple devices simultaneously.
- **Device Management:** Save, load, and manage your device list.
- **Cross-Platform:** Build and run on macOS, Linux, and Windows.

## 🚀 Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install)
- [Fyne](https://developer.fyne.io/started/)

### Installation & Running

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/ispapp/psshclient.git
    cd psshclient
    ```

2.  **Run the application:**
    ```sh
    make run
    ```

## 📦 Building for Production

Create distributable packages for different operating systems:

- **macOS:**
  ```sh
  make build-mac
  ```

- **Linux:**
  ```sh
  make build-linux
  ```

- **Windows:**
  ```sh
  make build-win
  ```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
