class ExtTags < Formula
  desc "Write and import foobar2000 external .tag files"
  homepage "https://github.com/daregod/ext-tags"
  version "0.11.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/daregod/ext-tags/releases/download/v0.11.1/ext-tags_0.11.1_darwin_arm64.tar.gz"
      sha256 "3b020282f0b3eadda145864ecb7209cfda8cc7010a9951b9a3789742cade876d"
    else
      url "https://github.com/daregod/ext-tags/releases/download/v0.11.1/ext-tags_0.11.1_darwin_amd64.tar.gz"
      sha256 "5e15eb74f52742462189d4d98ca0e434c2292c7966afcb0589ff08f8397647b7"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/daregod/ext-tags/releases/download/v0.11.1/ext-tags_0.11.1_linux_arm64.tar.gz"
      sha256 "72fbc7985fa8945c9ef2cb8d8b9093bc0ed62caadcf10162af5dca35add41e55"
    else
      url "https://github.com/daregod/ext-tags/releases/download/v0.11.1/ext-tags_0.11.1_linux_amd64.tar.gz"
      sha256 "edb985a0fb5e8548662e5af5f66c43c244cf2bf35beeca6c0ef19e79e82507fe"
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
