require 'rubygems'

$commands = "cmd/main.go cmd/build.go cmd/stack.go cmd/apps.go " +
            "cmd/deploy.go cmd/kubectl.go cmd/env.go cmd/logs.go " + 
            "cmd/helper.go cmd/run.go cmd/infra.go cmd/upgrade.go"

$bucket_prefix = "gs://datacol-distros"

$bin_matrix = {
  darwin: ['386', 'amd64'],
  linux: ['arm', '386', 'amd64'],
  # windows: ['386', 'amd64']
}

$version = ENV.fetch('VERSION')
$cmd_name = "datacol"
$api_name = "apictl"

def build_all
  $bin_matrix.each do |os, archs|
    with_cmd("mkdir -p dist/#{$version}")
    archs.each do |arch|
      bin_name = "#{$cmd_name}-#{os}-#{arch}"
      bin_name += ".exe" if os == 'windows'

      with_cmd("GOOS=#{os} GOARCH=#{arch} go build -o dist/#{$version}/#{bin_name} #{$commands}")
    end
  end
end


def apictl
  api_name = "apictl"
  os, arch = 'linux', 'amd64'
  with_cmd("GOOS=#{os} GOARCH=#{arch} go build -o dist/#{$version}/#{api_name} api/*.go")

  binary_dest = "#{$bucket_prefix}/binaries/#{$version}/#{api_name}.zip"
  version_dir = "dist/#{$version}"

  with_cmd("pushd #{version_dir} && \
            zip #{api_name}.zip #{api_name} && \
            gsutil cp #{api_name}.zip #{binary_dest} && \
            gsutil acl ch -u AllUsers:R #{binary_dest}
            popd")
end

def push_all
  build_all
  binary_dir = "#{$bucket_prefix}/binaries"
  latest_txt_path = "#{binary_dir}/latest.txt"
  version_dir = "dist/#{$version}"

  push_zip

  with_cmd("gsutil cp -r #{version_dir} #{binary_dir}")
  with_cmd("echo #{$version} > dist/latest.txt")
  with_cmd("gsutil cp dist/latest.txt #{binary_dir}/latest.txt")
  with_cmd("gsutil acl ch -u AllUsers:R -r #{binary_dir}")
end

def push_zip
  version_dir = "dist/#{$version}"

  { osx: 'darwin-386', linux: 'linux-amd64' }.each do |zipbin, name|
    with_cmd("pushd #{version_dir} && \
             cp #{$cmd_name}-#{name} datacol && \
             zip #{zipbin}.zip datacol && \
             gsutil cp #{zipbin}.zip #{$bucket_prefix}/ && \
             popd")

    with_cmd("gsutil acl ch -u AllUsers:R #{$bucket_prefix}/#{zipbin}.zip")
  end
end

def with_cmd(cmd)
  puts "#{cmd}"
  `#{cmd}`
  if code = $?.exitstatus > 0
    puts "exiting b/c of error"
    exit(code)
  end
end

if ARGV.size > 0
  send(ARGV[0])
else
  push_all
end