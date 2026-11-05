class ExtTags < Formula
  desc "Write and import foobar2000 external .tag files"
  homepage "https://github.com/daregod/ext-tags"
  version "0.11.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/daregod/ext-tags/releases/download/v0.11.0/ext-tags_0.11.0_darwin_arm64.tar.gz"
      sha256 "9007e52c20a8a699d9cde8cc04ef697a42136aa21728634d8dd4ee8437dbd519"
    else
      url "https://github.com/daregod/ext-tags/releases/download/v0.11.0/ext-tags_0.11.0_darwin_amd64.tar.gz"
      sha256 "e60a2ded15f4c8e474efdbce742216415e02e4e4da4205b32329143ee8dfd1b9"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/daregod/ext-tags/releases/download/v0.11.0/ext-tags_0.11.0_linux_arm64.tar.gz"
      sha256 "301f8f5e48ec767b999e11449c3d6d9d39e3c21ef6abcf6f142b4ccd20e0abf1"
    else
      url "https://github.com/daregod/ext-tags/releases/download/v0.11.0/ext-tags_0.11.0_linux_amd64.tar.gz"
      sha256 "d7869cec134c6f5ba4a0606544f32acc7811a74bbe25c44fc7c951b5595e9312"
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
