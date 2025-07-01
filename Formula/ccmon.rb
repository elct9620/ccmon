class Ccmon < Formula
  desc "TUI application for monitoring Claude Code API usage through OpenTelemetry"
  homepage "https://github.com/elct9620/ccmon"
  license "Apache-2.0"
  version "0.4.0"

  on_macos do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_x86_64.tar.gz"
      sha256 "fd7f35d245c5dd94363aae170274db4d4f2b16f697b6a5a646afeb7f015a4959"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_arm64.tar.gz"
      sha256 "14ba1a43d0f2dce87f5913edeea7acc4868510551de9103f8dacd2ddfe65d29c"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_x86_64.tar.gz"
      sha256 "dffad5754eb6f62e8925fba018438d1c739bd406ea20401dcb8e3fc8d29defd8"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_arm64.tar.gz"
      sha256 "40223baa703f30ec87be2f6a6a96b1af3717d2a812ef0230b678286c5b0f2f33"
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
