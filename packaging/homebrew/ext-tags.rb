class ExtTags < Formula
  desc "Write and import foobar2000 external .tag files"
  homepage "https://github.com/daregod/ext-tags"
  version "0.10.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.1/ext-tags_0.10.1_darwin_arm64.tar.gz"
      sha256 "5235010a6eaef6c58d4613a70d094fd1c6580fadc54907f5c73c8af28c1d1688"
    else
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.1/ext-tags_0.10.1_darwin_amd64.tar.gz"
      sha256 "d5b0c3b3a15b1397178e090af931f5c2bc44c0a6a08c09c503ba9f215919e88c"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.1/ext-tags_0.10.1_linux_arm64.tar.gz"
      sha256 "b82c1200d14e5280eb8ac6af9a475a084e6abb8cca26674735206efa69b6893d"
    else
      url "https://github.com/daregod/ext-tags/releases/download/v0.10.1/ext-tags_0.10.1_linux_amd64.tar.gz"
      sha256 "559d31b763832b43c37748ee843c89f4fbb72fa7d5c0b9ad4afb14617d77f5b8"
    end
  end

  def install
    binary = Dir["ext-tags_*"].first
    odie "ext-tags binary not found" if binary.nil?
    bin.install binary => "ext-tags"
  end

  test do
    assert_match "ext-tags", shell_output("#{bin}/ext-tags version")
  end
end
