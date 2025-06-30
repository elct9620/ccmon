class Ccmon < Formula
  desc "TUI application for monitoring Claude Code API usage through OpenTelemetry"
  homepage "https://github.com/elct9620/ccmon"
  license "Apache-2.0"
  version "0.3.2"

  on_macos do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_x86_64.tar.gz"
      sha256 "9bda6c6b66637732a94a7ac508f1a67b6c90ec614121c54be2579cf7fd6a7620"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_arm64.tar.gz"
      sha256 "0ea9538f103ee08ba0c5b180a6d8c70651ffc9ce9880231d48043f83b8bcdd36"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_x86_64.tar.gz"
      sha256 "9a2bb247e33e6853ce1d42cde2ff7ce0cbb8c5339d3a923afceab51618ce3414"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_arm64.tar.gz"
      sha256 "994a3f3450e17e73cc27e983cf08895c9a35c5177f08303639a9634ae4b8a654"
    end
  end

  head do
    url "https://github.com/elct9620/ccmon.git", branch: "main"

    depends_on "go" => :build
    depends_on "protobuf" => :build
  end

  def install
    if build.head?
      # Build from source for --head installations
      # Set up GOPATH and install required Go tools for protobuf generation
      ENV["GOPATH"] = buildpath/"go"
      ENV["PATH"] = "#{buildpath}/go/bin:#{ENV.fetch("PATH")}"

      system "go", "install", "google.golang.org/protobuf/cmd/protoc-gen-go@latest"
      system "go", "install", "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"

      # Generate protobuf code
      system "make", "generate"

      # Build the binary
      system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=HEAD"), "."
    else
      # Install pre-built binary for stable releases
      bin.install "ccmon"
    end
  end

  service do
    run [opt_bin/"ccmon", "--server"]
    keep_alive false
    working_dir var
    log_path var/"log/ccmon.log"
    error_log_path var/"log/ccmon.log"
  end

  test do
    # Test basic functionality
    system bin/"ccmon", "--help"

    # Test server mode startup (should exit cleanly with --help after server flag)
    system bin/"ccmon", "--server", "--help"

    # Test that binary executes without crashing
    output = shell_output("#{bin}/ccmon --help 2>&1", 0)
    assert_match(/Usage of.*ccmon/, output)
  end
end
