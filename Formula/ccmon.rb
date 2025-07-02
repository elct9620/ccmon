class Ccmon < Formula
  desc "TUI application for monitoring Claude Code API usage through OpenTelemetry"
  homepage "https://github.com/elct9620/ccmon"
  license "Apache-2.0"
  version "0.6.0"

  on_macos do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_x86_64.tar.gz"
      sha256 "b0aa22a5b2f717ce3d5f537db7f0c8430000c35b000ba1cb7d676b700f851868"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_arm64.tar.gz"
      sha256 "53f270ca1a4049da71be278b4d9aca43c0e99d80858241c69c5bc9f9537f69aa"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_x86_64.tar.gz"
      sha256 "d115543456d284074626e0c20b371fcf604183e6847676784b73245e5b78f2f8"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_arm64.tar.gz"
      sha256 "2bd6a9dac41af06f68995b4168cfe96f55cd3f500c09f0490d4469f8dba51434"
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
