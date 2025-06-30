class Ccmon < Formula
  desc "TUI application for monitoring Claude Code API usage through OpenTelemetry"
  homepage "https://github.com/elct9620/ccmon"
  license "Apache-2.0"
  version "0.2.1"

  on_macos do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_x86_64.tar.gz"
      sha256 "PLACEHOLDER_INTEL_SHA256"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM_SHA256"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_x86_64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_INTEL_SHA256"
    end

    on_arm do
      url "https://github.com/elct9620/ccmon/releases/download/v#{version}/ccmon_Linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_ARM_SHA256"
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
      system "make", "generate"
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

    # Test version output (works for both stable and HEAD builds)
    output = shell_output("#{bin}/ccmon --version 2>&1", 0)
    if build.head?
      assert_match "HEAD", output
    else
      assert_match version.to_s, output
    end
  end
end
