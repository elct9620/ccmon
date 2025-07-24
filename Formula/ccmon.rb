class Ccmon < Formula
  desc "TUI application for monitoring Claude Code API usage through OpenTelemetry"
  homepage "https://github.com/elct9620/ccmon"
  license "Apache-2.0"
  version "0.9.2"

  on_macos do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_x86_64.tar.gz"
      sha256 "5b178ad4f318999cdc02f54f15cebe89cc82faffdecba5a08166123480dc0f6a"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_arm64.tar.gz"
      sha256 "24890743f6d92a9a81b75dbdcd50e07805d3daab1995ab5c039b180050bf355a"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_x86_64.tar.gz"
      sha256 "6ebf94bd024b0e8ee5d845908a3bed774dcbc637a06a01d6519cdec4b57af8da"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_arm64.tar.gz"
      sha256 "8ca218522b8156dee72153bd5d1e899cb829262e3468592eb559695660f4ed31"
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
