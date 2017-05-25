# SCOIR Technical Interview for Back-End Engineers
This repo contains an exercise intended for Back-End Engineers.

## Instructions
1. Fork this repo.
1. Using technology of your choice, complete [the assignment](./Assignment.md).
1. Update this README with
    * a `How-To` section containing any instructions needed to execute your program.
    * an `Assumptions` section containing documentation on any assumptions made while interpreting the requirements.
1. Before the deadline, submit a pull request with your solution.

## Expectations
1. Please take no more than 8 hours to work on this exercise. Complete as much as possible and then submit your solution.
1. This exercise is meant to showcase how you work. With consideration to the time limit, do your best to treat it like a production system.

## How-To
### Build
* `make` will pull dependencies and build the binary
```
make
```

### Run 
* To run using the in, out, err, etc. directories included in the repository, run 
```
./run_with_defaults
```
* To run using the same as above, but also move some testing data into the `in` directory:
```
mv test_data/* in/*`, run `./run_with_defaults test_data
```
* To run using directories you specify:
```
bin/fileconverter -alsologtostderr -completed $PWD/done -errors $PWD/err -input $PWD/in -output $PWD/out -log_dir $PWD/log
```

## Assumptions
1. Required: Go 1.8+
1. Required: GNU Make compatible with GNU Make 3.82
1. Instead of deleting completed files, better to move input files to a `completed` directory rather than delete them.  They'll be needed to examine for line level errors.
1. Input files will be completely written to disk elsewhere on the same volume, closed, and moved into the `input-directory` in an atomic fashion, i.e. mv command or rename(). 
1. Regarding the requirement, "files will be considered new if the file name has not been recorded as processed before".  In order to implement this correctly, we will need to implement some data store of processed file names which needs to be persisted between application runs.  Assumption: for this exercise, it is sufficient to check the `output-directory` for a file with the same name.  In a production system, the output directory would get cleaned up occassionally so that may not be sufficient. Although, you may want reprocess files at a later time and simply removing the output file would be a convenient way of doing that.
1. Although, `MIDDLE_NAME` is optional, input files must contain a field for this value.
1. Input files are simple csv as shown in the example, and NOT quoted strings like: 12345,"Ryan",,"McCann","123-111-1111"
1. Assuming that 8 digit `INTERNAL_ID` can be 1 to 8 digits
1. Field values cannot contain commas or new lines
1. Assuming single byte characters for simplicy sake.  Otherwise, name validation of the 15 character limit has to count runes and handle special cases. 
