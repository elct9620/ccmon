class Ccmon < Formula
  desc "TUI application for monitoring Claude Code API usage through OpenTelemetry"
  homepage "https://github.com/elct9620/ccmon"
  license "Apache-2.0"
  version "0.9.0"

  on_macos do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_x86_64.tar.gz"
      sha256 "9c01cad2e988184a9cd58ff028f77a3041563f7c23c7ab3a2b6eb8d151a5afff"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_arm64.tar.gz"
      sha256 "6df25702e37340a9cfe96c764299d819c040ea94a3d569e6232cbe7a37bb8f6e"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_x86_64.tar.gz"
      sha256 "ae346ad0256285ccbb15a4579fc55bf6312ee8251ea5c79ab7bc35593ea7bce2"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_arm64.tar.gz"
      sha256 "d5b796ab78f1be75c482455bf9ca374d7e6e562efad22183e13b09ceae31cefc"
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
